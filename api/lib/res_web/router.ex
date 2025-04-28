defmodule ResWeb.Router do
  use ResWeb, :router

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/", ResWeb do
    pipe_through :api
    get "/images/", ImageController, :index
    post "/images/", ImageController, :create
    get "/images/:name", ImageController, :link_show
    delete "/images/:name", ImageController, :delete
    get "/images/:name/properties", ImageController, :show
    patch "/images/:name/properties", ImageController, :update
    put "/images/:name/raw", ImageController, :upload
  end
end
