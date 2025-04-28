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

    case image do
      nil ->
        {:error, :not_found}

      _ ->
        render(conn, :show, image: image)
    end
  end

  def update(conn, %{"name" => name} = params) do
    image = Images.get_image_by_name!(name)

    if image == nil do
      {:error, :not_found}
    end

    with {:ok, %Image{} = image} <- Images.update_image(image, params) do
      render(conn, :show, image: image)
    end
  end

  def delete(conn, %{"name" => name}) do
    image = Images.get_image_by_name!(name)

    case image do
      nil ->
        {:error, :not_found}

      _ ->
        with {:ok, %Image{}} <- Images.delete_image(image) do
          send_resp(conn, :no_content, "")
        end
    end
  end

  def link_show(conn, %{"name" => name}) do
    conn
    |> put_status(:moved_permanently)
    |> put_resp_header("location", ~p"/images/#{name}/properties")
    |> send_resp(:moved_permanently, "")
  end

  def upload(conn, %{"name" => name, "data" => %Plug.Upload{path: path, filename: filename}}) do
    image = Images.get_image_by_name!(name)

    case image do
      nil ->
        {:error, :not_found}

      _ ->
        base_dir = Application.get_env(:res, :base_dir)
        image_dir = Path.join([base_dir, "images", name])
        File.mkdir_p!(image_dir)
        image_file = Path.join([image_dir, filename])
        File.cp!(path, image_file)

        image_file =
          case Path.extname(filename) do
            ".zst" ->
              {output, status} = System.cmd("zstd", ["-d", "-c", path])

              if status == 0 do
                decompressed_filename = String.replace(filename, ".zst", "")
                decompressed_path = Path.join([image_dir, decompressed_filename])
                File.rename!(image_file, decompressed_path)
                decompressed_path
              else
                raise "Failed to decompress file: #{output}"
              end

            _ ->
              image_file
          end

        with {:ok, %Image{} = image} <-
               Images.get_image_by_name!(name)
               |> Images.update_image_raw(%{
                 "path" => image_file,
                 "size" => File.stat!(image_file).size,
                 "status" => "uploaded",
                 "format" => Path.extname(filename)
               }) do
          conn
          |> put_status(:ok)
          |> put_resp_header("location", ~p"/images/#{name}/properties")
          |> render(:show, image: image)
        end
    end
  end
end
