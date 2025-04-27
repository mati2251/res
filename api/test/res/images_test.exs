defmodule Res.ImagesTest do
  use Res.DataCase

  alias Res.Images

  describe "images" do
    alias Res.Images.Image

    import Res.ImagesFixtures

    @invalid_attrs %{name: nil, size: nil, status: nil, path: nil, format: nil, description: nil}

    test "list_images/0 returns all images" do
      image = image_fixture()
      assert Images.list_images() == [image]
    end

    test "get_image!/1 returns the image with given id" do
      image = image_fixture()
      assert Images.get_image!(image.id) == image
    end

    test "create_image/1 with valid data creates a image" do
      valid_attrs = %{name: "some name", size: 42, status: "some status", path: "some path", format: "some format", description: "some description"}

      assert {:ok, %Image{} = image} = Images.create_image(valid_attrs)
      assert image.name == "some name"
      assert image.size == 42
      assert image.status == "some status"
      assert image.path == "some path"
      assert image.format == "some format"
      assert image.description == "some description"
    end

    test "create_image/1 with invalid data returns error changeset" do
      assert {:error, %Ecto.Changeset{}} = Images.create_image(@invalid_attrs)
    end

    test "update_image/2 with valid data updates the image" do
      image = image_fixture()
      update_attrs = %{name: "some updated name", size: 43, status: "some updated status", path: "some updated path", format: "some updated format", description: "some updated description"}

      assert {:ok, %Image{} = image} = Images.update_image(image, update_attrs)
      assert image.name == "some updated name"
      assert image.size == 43
      assert image.status == "some updated status"
      assert image.path == "some updated path"
      assert image.format == "some updated format"
      assert image.description == "some updated description"
    end

    test "update_image/2 with invalid data returns error changeset" do
      image = image_fixture()
      assert {:error, %Ecto.Changeset{}} = Images.update_image(image, @invalid_attrs)
      assert image == Images.get_image!(image.id)
    end

    test "delete_image/1 deletes the image" do
      image = image_fixture()
      assert {:ok, %Image{}} = Images.delete_image(image)
      assert_raise Ecto.NoResultsError, fn -> Images.get_image!(image.id) end
    end

    test "change_image/1 returns a image changeset" do
      image = image_fixture()
      assert %Ecto.Changeset{} = Images.change_image(image)
    end
  end
end
