defmodule Res.Images.Image do
  use Ecto.Schema
  import Ecto.Changeset

  schema "images" do
    field :name, :string
    field :size, :integer
    field :status, :string
    field :path, :string
    field :format, :string
    field :description, :string

    timestamps(type: :utc_datetime)
  end

  @doc false
  def changeset(image, attrs) do
    image
    |> cast(attrs, [:name, :description, :size, :path, :status, :format])
    |> validate_required([:name, :description, :status])
    |> unique_constraint(:name, name: :unique_name)
  end
end
