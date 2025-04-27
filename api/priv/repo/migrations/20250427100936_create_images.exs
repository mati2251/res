defmodule Res.Repo.Migrations.CreateImages do
  use Ecto.Migration

  def change do
    create table(:images) do
      add :name, :string
      add :description, :string
      add :size, :integer
      add :path, :string
      add :status, :string
      add :format, :string

      timestamps(type: :utc_datetime)
    end
  end
end
