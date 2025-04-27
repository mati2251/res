defmodule ResWeb.ImageControllerTest do
  use ResWeb.ConnCase

  import Res.ImagesFixtures

  alias Res.Images.Image

  @create_attrs %{
    name: "some name",
    size: 42,
    status: "some status",
    path: "some path",
    format: "some format",
    description: "some description"
  }
  @update_attrs %{
    name: "some updated name",
    size: 43,
    status: "some updated status",
    path: "some updated path",
    format: "some updated format",
    description: "some updated description"
  }
  @invalid_attrs %{name: nil, size: nil, status: nil, path: nil, format: nil, description: nil}

  setup %{conn: conn} do
    {:ok, conn: put_req_header(conn, "accept", "application/json")}
  end

  describe "index" do
    test "lists all images", %{conn: conn} do
      conn = get(conn, ~p"/api/images")
      assert json_response(conn, 200)["data"] == []
    end
  end

  describe "create image" do
    test "renders image when data is valid", %{conn: conn} do
      conn = post(conn, ~p"/api/images", image: @create_attrs)
      assert %{"id" => id} = json_response(conn, 201)["data"]

      conn = get(conn, ~p"/api/images/#{id}")

      assert %{
               "id" => ^id,
               "description" => "some description",
               "format" => "some format",
               "name" => "some name",
               "path" => "some path",
               "size" => 42,
               "status" => "some status"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn} do
      conn = post(conn, ~p"/api/images", image: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "update image" do
    setup [:create_image]

    test "renders image when data is valid", %{conn: conn, image: %Image{id: id} = image} do
      conn = put(conn, ~p"/api/images/#{image}", image: @update_attrs)
      assert %{"id" => ^id} = json_response(conn, 200)["data"]

      conn = get(conn, ~p"/api/images/#{id}")

      assert %{
               "id" => ^id,
               "description" => "some updated description",
               "format" => "some updated format",
               "name" => "some updated name",
               "path" => "some updated path",
               "size" => 43,
               "status" => "some updated status"
             } = json_response(conn, 200)["data"]
    end

    test "renders errors when data is invalid", %{conn: conn, image: image} do
      conn = put(conn, ~p"/api/images/#{image}", image: @invalid_attrs)
      assert json_response(conn, 422)["errors"] != %{}
    end
  end

  describe "delete image" do
    setup [:create_image]

    test "deletes chosen image", %{conn: conn, image: image} do
      conn = delete(conn, ~p"/api/images/#{image}")
      assert response(conn, 204)

      assert_error_sent 404, fn ->
        get(conn, ~p"/api/images/#{image}")
      end
    end
  end

  defp create_image(_) do
    image = image_fixture()
    %{image: image}
  end
end
