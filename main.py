from fastapi import FastAPI
from routers import images

app = FastAPI(title="Remote Script Executor API")

app.include_router(images.router, prefix="/images", tags=["images"])
