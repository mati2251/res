defmodule ResWeb.ImageController do
  use ResWeb, :controller

  alias Res.Images
  alias Res.Images.Image

  action_fallback ResWeb.FallbackController

  def index(conn, _params) do
    images = Images.list_images()
    render(conn, :index, images: images)
  end

  def create(conn, %{"name" => name} = image_params) do
    image_params = Map.put(image_params, "status", "created")
    with {:ok, %Image{} = image} <- Images.create_image(image_params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", ~p"/images/#{name}")
      |> render(:show, image: image)
    end
  end

  def show(conn, %{"name" => name}) do
    image = Images.get_image_by_name!(name)
    render(conn, :show, image: image)
  end

  def update(conn, %{"id" => id, "image" => image_params}) do
    image = Images.get_image!(id)

    with {:ok, %Image{} = image} <- Images.update_image(image, image_params) do
      render(conn, :show, image: image)
    end
  end

  def delete(conn, %{"id" => id}) do
    image = Images.get_image!(id)

    with {:ok, %Image{}} <- Images.delete_image(image) do
      send_resp(conn, :no_content, "")
    end
  end
end
