import os
import subprocess
import zipfile
from io import BytesIO
import logging
from pydantic import BaseModel

logger = logging.getLogger("uvicorn.error")

JOBS_STORE = ".store/jobs"
IMAGES_STORE = ".store/images"
ROOT_MOUNT = "root"
PROPERTIES_NAME = "properties"
OVERLAY_DIR = "overlay"
LOG_FILE = "job.log"
IMAGE_NAME = "image.sif"
SCRIPT_NAME = "script"
IMAGE_ATTR = "user.image"
EXIT_CODE_ATTR = "user.exit_code"
STATE_ATTR = "user.state"

os.makedirs(JOBS_STORE, exist_ok=True)


class Image(BaseModel):
    id: int
    state: str
    script: str
    exit_code: int
    image: str


class LaunchException(Exception):
    def __init__(self, message: str):
        super().__init__(message)
        self.message = message


class JobException(Exception):
    def __init__(self, message: str):
        super().__init__(message)
        self.message = message


class ImageProperties(BaseModel):
    image: str
    artifacts: list[str] | None = None


class FileProperties(BaseModel):
    name: str
    size: str
    type: str


def create_job() -> int:
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
    return max_job_id


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


def update_job(job_id: int, props: ImageProperties):
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        raise JobException("Job not found")

    image_path = os.path.abspath(f"{IMAGES_STORE}/{props.image}.sif")
    if not os.path.exists(image_path):
        raise JobException("Image not found")

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
    return job


def put_script(job_id: int, script_content: bytes):
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        raise JobException("Job not found")

    script_path = f"{job_path}/{SCRIPT_NAME}"
    with open(script_path, "w") as f:
        f.write(script_content.decode())

    os.chmod(script_path, 0o755)

    return job_from_id(job_id)


def get_script(job_id: int) -> str:
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        raise JobException("Job not found")

    script_path = f"{job_path}/{SCRIPT_NAME}"
    if not os.path.exists(script_path):
        raise JobException("Script not found")

    with open(script_path, "r") as f:
        return f.read().strip()


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


def set_state(job_id: int, state: str):
    """
    Set the job state to 'start' or 'done'.
    """
    job_path = f"{JOBS_STORE}/{job_id}"
    if not os.path.exists(job_path):
        raise JobException("Job not found")

    os.setxattr(job_path, STATE_ATTR, f"{state}ed".encode(), follow_symlinks=False)

    return state


def get_log(job_id: int) -> str:
    """
    Get the job log.
    """
    log_path = f"{JOBS_STORE}/{job_id}/{LOG_FILE}"
    if not os.path.exists(log_path):
        raise JobException("Log file not found")

    with open(log_path, "r") as f:
        return f.read()


def get_artifacts(job_id: int) -> list[FileProperties]:
    artifacts_path = f"{JOBS_STORE}/{job_id}/{ROOT_MOUNT}"

    if not os.path.exists(artifacts_path):
        raise JobException("Artifacts not found")

    artifacts_raw = ""

    try:
        artifacts_raw = os.getxattr(
            f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}",
            "user.artifacts",
            follow_symlinks=False,
        ).decode()
    except OSError:
        raise JobException("No artifacts found")

    artifacts = artifacts_raw.split(",")

    files = []
    for artifact in artifacts:
        artifact_path = os.path.abspath(os.path.join(artifacts_path, artifact))
        if os.path.exists(artifact_path):
            file_info = FileProperties(
                name=artifact,
                size=str(os.path.getsize(artifact_path)),
                type=subprocess.getoutput(
                    f"file -b --mime-type '{artifact_path}'"
                ).strip(),
            )
            files.append(file_info)

    return files


def get_artifacts_raw(job_id: int) -> BytesIO:
    artifacts_path = f"{JOBS_STORE}/{job_id}/{ROOT_MOUNT}"
    if not os.path.exists(artifacts_path):
        raise JobException("Artifacts not found")

    artifacts_raw = ""
    try:
        artifacts_raw = os.getxattr(
            f"{JOBS_STORE}/{job_id}/{PROPERTIES_NAME}",
            "user.artifacts",
            follow_symlinks=False,
        ).decode()
    except OSError:
        raise JobException("No artifacts found")

    artifacts = artifacts_raw.split(",")

    zip_buffer = BytesIO()
    with zipfile.ZipFile(zip_buffer, "w", zipfile.ZIP_DEFLATED) as zip_file:
        for artifact in artifacts:
            artifact_path = os.path.join(artifacts_path, artifact)
            if os.path.exists(artifact_path):
                zip_file.write(artifact_path, arcname=artifact)

    zip_buffer.seek(0)
    return zip_buffer


def get_jobs(state: str) -> list[Image]:
    jobs = []
    for job_id in os.listdir(JOBS_STORE):
        if job_id.isdigit():
            job = job_from_id(int(job_id))
            if job:
                if state == "" or state.lower() in job.state:
                    jobs.append(job)
    return jobs
