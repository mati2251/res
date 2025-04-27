defmodule ResWeb.ImageJSON do
  alias Res.Images.Image

  @doc """
  Renders a list of images.
  """
  def index(%{images: images}) do
    %{data: for(image <- images, do: data(image))}
  end

  @doc """
  Renders a single image.
  """
  def show(%{image: image}) do
    %{data: data(image)}
  end

  defp data(%Image{} = image) do
    %{
      id: image.id,
      name: image.name,
      description: image.description,
      size: image.size,
      path: image.path,
      status: image.status,
      format: image.format
    }
  end
end
