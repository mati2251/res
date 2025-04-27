defmodule ResWeb.Router do
  use ResWeb, :router

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/", ResWeb do
    pipe_through :api
    get "/images/", ImageController, :index
    post "/images/", ImageController, :create
    get "/images/:name", ImageController, :show
  end
end
