from dataclasses import dataclass
import logging
import os
import subprocess
import zipfile
from io import BytesIO

from fastapi import APIRouter, File, UploadFile
from fastapi.responses import JSONResponse, PlainTextResponse
from pydantic import BaseModel

logger = logging.getLogger("uvicorn.error")

router = APIRouter()

JOBS_STORE = ".store/jobs"
IMAGES_STORE = ".store/images"
ROOT_MOUNT = "/root"
PROPERTIES_NAME = "properties"
OVERLAY_DIR = "overlay"
LOG_FILE = "job.log"
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


def job_state(job_id: int) -> str:
    state = "not ready"
    image_path = f"{JOBS_STORE}/{job_id}/{IMAGE_NAME}"
    if not os.path.exists(image_path):
        return state
    script_path = f"{JOBS_STORE}/{job_id}/{SCRIPT_NAME}"
    if not os.path.exists(script_path):
        return state
    state = "ready"
    try:
        state = os.getxattr(
            f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}",
            STATE_ATTR,
            follow_symlinks=False,
        ).decode()
    except OSError:
        pass
    return state


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

    state = job_state(job_id)

    properties_path = f"{job_path}/{PROPERTIES_NAME}"

    try:
        exit_code = int(
            os.getxattr(properties_path, EXIT_CODE_ATTR, follow_symlinks=False).decode()
        )
    except OSError:
        exit_code = -1

    return Image(
        id=job_id, state=state, script=script, exit_code=exit_code, image=image
    )


class ImageProperties(BaseModel):
    image: str
    artifacts: list[str] | None = None


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
    if props.artifacts:
        os.setxattr(
            properties_path,
            "user.artifacts",
            ",".join(props.artifacts).encode(),
            follow_symlinks=False,
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
        os.chmod(script_path, 0o755)

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
        headers={"Location": f"/jobs/{job_id}/script/"},
    )


@router.get("/{job_id}/state/", response_model=str)
def get_state(job_id: int):
    """
    Get the state of a job.
    """
    state = job_state(job_id)
    return PlainTextResponse(
        status_code=200, content=state, headers={"Location": f"/jobs/{job_id}/state/"}
    )


class LaunchException(Exception):
    def __init__(self, message: str):
        super().__init__(message)
        self.message = message


def launch_script(job_id):
    script_path = os.path.abspath(f"{JOBS_STORE}/{job_id}/{SCRIPT_NAME}")
    if not os.path.exists(script_path):
        raise LaunchException("Script not found")
    image_path = os.path.abspath(f"{JOBS_STORE}/{job_id}/{IMAGE_NAME}")
    if not os.path.exists(image_path):
        raise LaunchException("Image not found")
    overlay_path = os.path.abspath(f"{JOBS_STORE}/{job_id}/{OVERLAY_DIR}")
    os.makedirs(overlay_path, exist_ok=True)

    properties_path = os.path.abspath(f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}")
    root_mount = os.path.abspath(f"{JOBS_STORE}/{job_id}/{ROOT_MOUNT}")
    os.makedirs(root_mount, exist_ok=True)

    cmd = [
        f"apptainer exec -C --fakeroot --bind {script_path} --bind {root_mount}:/root/ --overlay {overlay_path} {image_path} {script_path};",
        f"setfattr --name {EXIT_CODE_ATTR} --value $? {properties_path};"
        f"setfattr --name {STATE_ATTR} --value done {properties_path};",
    ]

    cmd = " ".join(cmd)
    cmd = f"bash -c '{cmd}'"

    logger.info(f"Launching job {job_id} with command: {cmd}")

    log_file = os.path.abspath(f"{JOBS_STORE}/{job_id}/{LOG_FILE}")
    with open(log_file, "w") as log_f:
        subprocess.Popen(cmd, shell=True, stdout=log_f, stderr=log_f)


@router.put("/{job_id}/state/", response_model=str)
def put_state(job_id: int, state: str):
    """
    Set the state of a job. Available states to set are: "start", "stop"
    """
    avalible_states = ["start", "stop"]
    if state not in avalible_states:
        return JSONResponse(
            status_code=400,
            content={"detail": f"State must be one of {avalible_states}"},
        )
    current_state = job_state(job_id)
    if current_state == "not ready":
        return JSONResponse(
            status_code=400, content={"detail": "Job is not ready to be started"}
        )
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        return JSONResponse(status_code=404, content={"detail": "Job not found"})

    os.setxattr(job_path, STATE_ATTR, f"{state}ed".encode(), follow_symlinks=False)

    if state == "start":
        try:
            launch_script(job_id)
        except LaunchException as e:
            return JSONResponse(status_code=500, content={"detail": str(e)})

    return PlainTextResponse(
        status_code=200, content=state, headers={"Location": f"/jobs/{job_id}/state/"}
    )


@router.get("/{job_id}/log/", response_model=str)
def get_logs(job_id: int):
    """
    Get job logs
    """
    log_path = f"{JOBS_STORE}/{job_id}/{LOG_FILE}"
    if not os.path.exists(log_path):
        return JSONResponse(status_code=404, content={"detail": "Log file not found"})

    with open(log_path, "r") as f:
        log_content = f.read()

    return PlainTextResponse(
        status_code=200,
        content=log_content,
        headers={"Location": f"/jobs/{job_id}/log/"},
    )

class FileProperties(BaseModel):
    name: str
    size: str
    type: str

@router.get("/{job_id}/artifacts/", response_model=list[FileProperties])
def get_artifacts(job_id: int):
    """
    Get job artifacts
    """
    artifacts_path = f"{JOBS_STORE}/{job_id}/{ROOT_MOUNT}"

    if not os.path.exists(artifacts_path):
        return JSONResponse(status_code=404, content={"detail": "Artifacts not found"})

    artifacts_raw = ""

    try:
        artifacts_raw = os.getxattr(
            f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}",
            "user.artifacts",
            follow_symlinks=False,
        ).decode()
    except OSError:
        return JSONResponse(status_code=404, content={"detail": "No artifacts found"})

    artifacts = artifacts_raw.split(",")

    files = []
    for artifact in artifacts:
        artifact_path = os.path.abspath(os.path.join(artifacts_path, artifact))
        if os.path.exists(artifact_path):
            file_info = FileProperties(
                name=artifact,
                size=str(os.path.getsize(artifact_path)),
                type=subprocess.getoutput(f"file -b --mime-type '{artifact_path}'").strip()
            )
            files.append(file_info)

    return files


@router.get("/{job_id}/artifacts/data")
def get_artifact_data(job_id: int):
    """
    Get archive of artifacts
    """
    artifacts_path = f"{JOBS_STORE}/{job_id}/{ROOT_MOUNT}"
    if not os.path.exists(artifacts_path):
        return JSONResponse(status_code=404, content={"detail": "Artifacts not found"})

    artifacts_raw = ""
    try:
        artifacts_raw = os.getxattr(
            f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}",
            "user.artifacts",
            follow_symlinks=False,
        ).decode()
    except OSError:
        return JSONResponse(status_code=404, content={"detail": "No artifacts found"})

    artifacts = artifacts_raw.split(",")

    zip_buffer = BytesIO()
    with zipfile.ZipFile(zip_buffer, "w", zipfile.ZIP_DEFLATED) as zip_file:
        for artifact in artifacts:
            artifact_path = os.path.join(artifacts_path, artifact)
            if os.path.exists(artifact_path):
                zip_file.write(artifact_path, arcname=artifact)
                
    zip_buffer.seek(0)
    return PlainTextResponse(
        status_code=200,
        content=zip_buffer.getvalue(),
        headers={
            "Content-Disposition": f'attachment; filename="artifacts_{job_id}.zip"',
            "Content-Type": "application/zip",
        },
    )
