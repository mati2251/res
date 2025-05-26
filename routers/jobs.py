import os
import logging

from fastapi import APIRouter, UploadFile, File
from fastapi.responses import JSONResponse, PlainTextResponse
from pydantic import BaseModel

logger = logging.getLogger("uvicorn.error")

router = APIRouter()

JOBS_STORE = ".store/jobs"
IMAGES_STORE = ".store/images"
PROPERTIES_NAME = "properties"
IMAGE_NAME = "image.sif"
SCRIPT_NAME = "script"
IMAGE_ATTR = "user.image"
EXIT_CODE_ATTR = "user.exit_code"
STATE_ATTR = "user.state"

os.makedirs(JOBS_STORE, exist_ok=True)


@router.post("/")
async def create_job():
    """
    Create a new job with the provided job data.
    """
    max_job_id = 0
    try:
        with open(f"{JOBS_STORE}/max_job_id.txt", "r") as f:
            max_job_id = int(f.read().strip())
    except FileNotFoundError:
        pass
    except ValueError:
        raise ValueError("Invalid job ID format in max_job_id.txt")
    max_job_id += 1
    with open(f"{JOBS_STORE}/max_job_id.txt", "w") as f:
        f.write(str(max_job_id))
    os.makedirs(f"{JOBS_STORE}/{max_job_id}", exist_ok=True)
    properties = f"{JOBS_STORE}/{max_job_id}/{PROPERTIES_NAME}"
    if not os.path.exists(properties):
        with open(properties, "w") as f:
            f.write("SEE EXTENDED ATTRIBUTES\n")
    return JSONResponse(
        status_code=201,
        content={
            "job_id": max_job_id,
        },
        headers={"Location": f"/jobs/{max_job_id}"},
    )


class Image(BaseModel):
    id: int
    state: str
    script: str
    exit_code: int
    image: str


def job_from_id(job_id: int) -> Image | None:
    """
    Get job properties from the job ID.
    """
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        return None
    try:
        image = os.getxattr(
            f"{job_path}/{PROPERTIES_NAME}", IMAGE_ATTR, follow_symlinks=False
        ).decode()
    except OSError:
        image = ""

    script = ""
    script_path = f"{job_path}/{SCRIPT_NAME}"
    if not os.path.exists(script_path):
        return Image(id=job_id, state="not ready", script="", exit_code=-1, image=image)
    with open(script_path, "r") as f:
        script = f.read().strip()

    state = "not ready"
    if image != "" and script != "":
        state = "ready"

    try:
        exit_code = int(
            os.getxattr(job_path, EXIT_CODE_ATTR, follow_symlinks=False).decode()
        )
    except OSError:
        exit_code = -1

    try:
        state = os.getxattr(job_path, STATE_ATTR, follow_symlinks=False).decode()
    except OSError:
        pass


    return Image(
        id=job_id, state=state, script=script, exit_code=exit_code, image=image
    )


class ImageProperties(BaseModel):
    image: str


@router.put("/{job_id}/properties", response_model=Image)
async def update_job(job_id: int, props: ImageProperties):
    """
    Update a job properties.
    """
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        return JSONResponse(status_code=404, content={"detail": "Job not found"})

    image_path = os.path.abspath(f"{IMAGES_STORE}/{props.image}.sif")
    if not os.path.exists(image_path):
        return JSONResponse(status_code=404, content={"detail": "Image not found"})

    symlink_path = f"{job_path}/{IMAGE_NAME}"
    if os.path.exists(symlink_path):
        os.remove(symlink_path)
    os.symlink(image_path, symlink_path)

    properties_path = f"{job_path}/{PROPERTIES_NAME}"
    os.setxattr(
        properties_path, IMAGE_ATTR, props.image.encode(), follow_symlinks=False
    )

    job = job_from_id(job_id)

    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})

    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}"},
    )

@router.get("/{job_id}/", response_model=Image)
async def get_job(job_id: int):
    """
    Get job properties by job ID.
    """
    job = job_from_id(job_id)
    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})
    
    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}"},
    )

@router.put("/{job_id}/script/", response_model=Image)
async def update_job_script(job_id: int, file: UploadFile = File(...)):
    """
    Update the script of a job.
    """
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        return JSONResponse(status_code=404, content={"detail": "Job not found"})
    script_path = f"{job_path}/{SCRIPT_NAME}"
    with open(script_path, "wb") as f:
        content = await file.read()
        f.write(content)

    job = job_from_id(job_id)
    if job is None:
        return JSONResponse(status_code=404, content={"detail": "Job not found"})
    return JSONResponse(
        status_code=200,
        content=job.model_dump(),
        headers={"Location": f"/jobs/{job_id}"},
    )

@router.get("/{job_id}/script/", response_model=str)
async def get_job_script(job_id: int):
    """
    Get the script of a job
    """
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        return JSONResponse(status_code=404, content={"detail": "Job not found"})
    script_path = f"{job_path}/{SCRIPT_NAME}"
    if not os.path.exists(script_path):
        return JSONResponse(status_code=404, content={"detail": "Script not found"})
    
    with open(script_path, "r") as f:
        script_content = f.read()
    
    return PlainTextResponse(
        status_code=200,
        content=script_content,
        headers={"Location": f"/jobs/{job_id}/script/"}
    )
