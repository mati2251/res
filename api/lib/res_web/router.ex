defmodule ResWeb.Router do
  use ResWeb, :router

  pipeline :api do
    plug :accepts, ["json"]
    plug :error_handler
  end

  scope "/", ResWeb do
    pipe_through :api
  end

  def error_handler(conn, _opts) do
    conn
    |> put_status(:internal_server_error)
    |> json(%{error: "Internal Server Error"})
  end
end
