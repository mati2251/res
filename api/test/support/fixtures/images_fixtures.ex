defmodule Res.ImagesFixtures do
  @moduledoc """
  This module defines test helpers for creating
  entities via the `Res.Images` context.
  """

  @doc """
  Generate a image.
  """
  def image_fixture(attrs \\ %{}) do
    {:ok, image} =
      attrs
      |> Enum.into(%{
        description: "some description",
        format: "some format",
        name: "some name",
        path: "some path",
        size: 42,
        status: "some status"
      })
      |> Res.Images.create_image()

    image
  end
end
