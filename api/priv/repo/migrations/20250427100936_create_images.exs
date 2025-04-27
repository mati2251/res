defmodule Res.Repo.Migrations.CreateImages do
  use Ecto.Migration

  def change do
    create table(:images) do
      add :name, :string, null: false
      add :description, :string, null: false
      add :size, :integer, null: true
      add :path, :string, null: true
      add :status, :string, null: false
      add :format, :string, null: true

      timestamps(type: :utc_datetime)
    end
    create unique_index(:images, [:name], name: :unique_name)
  end
end
