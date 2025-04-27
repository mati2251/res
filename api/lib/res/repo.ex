defmodule Res.Repo do
  use Ecto.Repo,
    otp_app: :res,
    adapter: Ecto.Adapters.Postgres
end
