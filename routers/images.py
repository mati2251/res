from fastapi import APIRouter, HTTPException, Request, UploadFile, File, Query
from fastapi.responses import FileResponse, Response
from fastapi.responses import JSONResponse
import os
import hashlib

from utils.job import HASH_ATTR

router = APIRouter()

IMAGE_STORE = ".store/images"

SUPPORTED_MEDIA_TYPES = ["application/json"]

if not os.path.exists(IMAGE_STORE):
    os.makedirs(IMAGE_STORE)


@router.put("/{name}/raw")
async def upload_image(name: str, request: Request, file: UploadFile = File(...)):
    """
    Upload a raw apptainer file.
    """
    if file.filename is None:
        raise HTTPException(status_code=400, detail="Filename is required")
    if not file.filename.endswith(".sif"):
        raise HTTPException(status_code=400, detail="File must be a .sif file")
    if file.content_type != "application/octet-stream":
        raise HTTPException(
            status_code=400, detail="File must be of type application/octet-stream"
        )

    file_path = f"{IMAGE_STORE}/{name}.sif"
    request_etag = request.headers.get("ETag")
    current_etag = ""
    try:
        current_etag = os.getxattr(file_path, HASH_ATTR).decode("utf-8")
    except OSError:
        pass

    if request_etag is None and current_etag is not None:
        return JSONResponse(
            status_code=428,
            headers={"Etag": current_etag},
            content={"detail": "Etag header is required for script update."},
        )
    elif request_etag and current_etag and request_etag != current_etag:
        return JSONResponse(
            status_code=412,
            content={"detail": "Etag mismatch. The script has been modified."},
        )

    with open(file_path, "wb") as f:
        content = await file.read()
        hash = hashlib.sha256(content).hexdigest()
        try:
            os.setxattr(file_path, HASH_ATTR, hash.encode("utf-8"))
        except OSError as e:
            raise HTTPException(status_code=500, detail=f"Error setting ETag: {str(e)}")
        f.write(content)
    if not content:
        raise HTTPException(status_code=400, detail="File is empty")

    return Response(status_code=201, headers={"Location": f"/images/{name}/properties"})


@router.get("/{name}/raw")
async def get_image_raw(name: str):
    """
    Get a raw apptainer image file (.sif).
    """
    file_path = os.path.join(IMAGE_STORE, f"{name}.sif")
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail=f"Image '{name}' not found")

    return FileResponse(
        path=file_path, media_type="application/octet-stream", filename=f"{name}.sif"
    )


@router.get("/{name}/properties")
def get_image_properties(name: str):
    """
    Get properties of an image.
    """
    file_path = os.path.join(IMAGE_STORE, f"{name}.sif")
    print(file_path)
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail=f"Image '{name}' not found")

    return {
        "name": name,
        "size": os.path.getsize(file_path),
        "type": "apptainer",
        "status": "available",
    }


@router.get("/{name}/")
def get_image(name: str):
    """
    Redirect to image properties endpoint.
    """
    return Response(status_code=303, headers={"Location": f"/images/{name}/properties"})


@router.delete("/{name}/")
async def delete_image(name: str):
    """
    Delete an image by name.
    """
    file_path = os.path.join(IMAGE_STORE, f"{name}.sif")
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail=f"Image '{name}' not found")

    os.remove(file_path)
    return Response(status_code=204)


@router.get("/")
def list_images(skip: int = Query(0, ge=0), limit: int = Query(10, gt=0)):
    """
    List all images in the store, with pagination support.
    """
    images = []
    for filename in sorted(os.listdir(IMAGE_STORE)):
        if filename.endswith(".sif"):
            name = filename[:-4]
            file_path = os.path.join(IMAGE_STORE, filename)
            images.append(
                {
                    "name": name,
                    "size": os.path.getsize(file_path),
                    "type": "apptainer",
                    "status": "available",
                }
            )

    paginated = images[skip : skip + limit]
    return {"total": len(images), "skip": skip, "limit": limit, "items": paginated}
