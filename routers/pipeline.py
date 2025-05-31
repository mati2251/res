import asyncio
import logging
from typing import List, Optional
from fastapi import APIRouter
from pydantic import BaseModel, Field
from fastapi.responses import JSONResponse

import utils.job as jobs


class PipelineJob(BaseModel):
    name: str
    image: str
    script: List[str]
    artifacts: Optional[List[str]] = Field(default_factory=list)


class PipelineDefinition(BaseModel):
    jobs: List[PipelineJob] = Field(default_factory=list)


router = APIRouter()
logger = logging.getLogger("uvicorn.error")


async def execute_pipeline(ids: List[int]):
    """
    Execute the pipeline by creating jobs and setting their states.
    """
    try:
        for i in range(len(ids)):
            jobs.set_state(ids[i], "started")
            await jobs.launch_and_wait(ids[i])
            if i < len(ids) - 1:
                next_job = ids[i + 1]
                jobs.cp_artifacts(ids[i], next_job)
    except Exception as e:
        logger.error(f"Error executing pipeline: {e}")


@router.post("/", response_model=List[int])
async def create_pipeline(pipeline: PipelineDefinition):
    """
    Create a new pipeline definition.
    """
    ids = []
    for job in pipeline.jobs:
        job_id = jobs.create_job()
        ids.append(job_id)
        properties = jobs.ImageProperties(image=job.image, artifacts=job.artifacts)
        try:
            jobs.update_job(job_id, properties)
            jobs.put_script(job_id, "\n".join(job.script).encode("utf-8"))
            jobs.set_state(job_id, "queued")
        except jobs.JobException as e:
            return JSONResponse(status_code=400, content={"detail": str(e)})

    asyncio.create_task(execute_pipeline(ids))
    return ids
