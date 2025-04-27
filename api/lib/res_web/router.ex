defmodule ResWeb.Router do
  use ResWeb, :router

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/", ResWeb do
    pipe_through :api
    post "/images/", ImageController, :create
  end
end
