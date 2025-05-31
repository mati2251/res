from fastapi import FastAPI
from routers import images, jobs, pipeline

app = FastAPI(title="Remote Script Executor API")

app.include_router(images.router, prefix="/images", tags=["images"])
app.include_router(jobs.router, prefix="/jobs", tags=["jobs"])
app.include_router(pipeline.router, prefix="/pipelines", tags=["pipeline"])
