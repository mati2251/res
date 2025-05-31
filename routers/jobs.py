import logging
import asyncio

from fastapi import APIRouter, UploadFile, File, Query, Request
from fastapi.responses import JSONResponse, PlainTextResponse
import utils.job as utils

logger = logging.getLogger("uvicorn.error")

router = APIRouter()


@router.post("/")
async def create_job():
    """
    Create a new job with the provided job data.
    """
    max_job_id = utils.create_job()
    return JSONResponse(
        status_code=201,
        content={
            "id": max_job_id,
        },
        headers={"Location": f"/jobs/{max_job_id}"},
    )


@router.put("/{job_id}/properties", response_model=utils.Image)
async def update_job(job_id: int, props: utils.ImageProperties):
    """
    Update a job properties.
    """
    job = None
    try:
        job = utils.update_job(job_id, props)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})

    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})

    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}"},
    )


@router.get("/{job_id}/", response_model=utils.Image)
async def get_job(job_id: int):
    """
    Get job properties by job ID.
    """
    job = utils.job_from_id(job_id)
    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})

    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}"},
    )


@router.put("/{job_id}/script/", response_model=utils.Image)
async def update_job_script(
    job_id: int, request: Request, file: UploadFile = File(...)
):
    """
    Update the script of a job.
    """
    request_etag = request.headers.get("Etag")
    script_etag = utils.get_script_etag(job_id)
    if request_etag is None and script_etag is not None:
        return JSONResponse(
            status_code=428,
            headers={"Etag": script_etag},
            content={"detail": "Etag header is required for script update."},
        )
    elif request_etag and script_etag and request_etag != script_etag:
        return JSONResponse(
            status_code=412,
            content={"detail": "Etag mismatch. The script has been modified."},
        )

    content = await file.read()
    job = None
    hash = None
    try:
        job, hash = utils.put_script(job_id, content)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})

    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})
    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}", "Etag": hash or ""},
    )


@router.get("/{job_id}/script/", response_model=str)
async def get_job_script(job_id: int):
    """
    Get the script of a job
    """
    content = ""
    try:
        content = utils.get_script(job_id)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})
    return PlainTextResponse(
        status_code=200,
        content=content,
        headers={"Location": f"/jobs/{job_id}/script/"},
    )


@router.get("/{job_id}/state/", response_model=str)
def get_state(job_id: int):
    """
    Get the state of a job.
    """
    state = utils.job_state(job_id)
    return PlainTextResponse(
        status_code=200, content=state, headers={"Location": f"/jobs/{job_id}/state/"}
    )


@router.put("/{job_id}/state/", response_model=str)
async def put_state(job_id: int, state: str):
    """
    Set the state of a job. Available states to set are: "start", "stop"
    """
    avalible_states = ["start", "stop"]
    if state not in avalible_states:
        return JSONResponse(
            status_code=400,
            content={"detail": f"State must be one of {avalible_states}"},
        )
    current_state = utils.job_state(job_id)
    if current_state == "not ready":
        return JSONResponse(
            status_code=400, content={"detail": "Job is not ready to be started"}
        )

    try:
        utils.set_state(job_id, state)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})

    if state == "start":
        try:
            asyncio.create_task(utils.launch_and_wait(job_id))
        except utils.LaunchException as e:
            return JSONResponse(status_code=500, content={"detail": str(e)})

    return PlainTextResponse(
        status_code=200, content=state, headers={"Location": f"/jobs/{job_id}/state/"}
    )


@router.get("/{job_id}/log/", response_model=str)
def get_logs(job_id: int):
    """
    Get job logs
    """
    log_content = ""
    try:
        log_content = utils.get_log(job_id)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})

    return PlainTextResponse(
        status_code=200,
        content=log_content,
        headers={"Location": f"/jobs/{job_id}/log/"},
    )


@router.get("/{job_id}/artifacts/", response_model=list[utils.FileProperties])
def get_artifacts(job_id: int):
    """
    Get job artifacts
    """
    artifacts = []
    try:
        artifacts = utils.get_artifacts(job_id)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})

    return artifacts


@router.get("/{job_id}/artifacts/data")
def get_artifact_data(job_id: int):
    """
    Get archive of artifacts
    """
    try:
        artifacts = utils.get_artifacts_raw(job_id)
    except utils.JobException as e:
        return JSONResponse(status_code=404, content={"detail": str(e)})
    return PlainTextResponse(
        status_code=200,
        content=artifacts.getvalue(),
        headers={
            "Content-Disposition": f'attachment; filename="artifacts_{job_id}.zip"',
            "Content-Type": "application/zip",
        },
    )


@router.get("/")
def list_jobs(
    skip: int = Query(0, ge=0), limit: int = Query(10, gt=0), state: str = Query("")
):
    """
    List all images in the store, with pagination support.
    """
    jobs = utils.get_jobs(state)
    if not jobs:
        return {"total": 0, "skip": skip, "limit": limit, "items": []}
    total = len(jobs)
    paginated = jobs[skip : skip + limit]
    return {
        "total": total,
        "skip": skip,
        "limit": limit,
        "items": [job.model_dump() for job in paginated],
    }
