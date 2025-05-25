from fastapi import Header, HTTPException

SUPPORTED_MEDIA_TYPES = ["application/json"]

def validate_accept_header(accept: str = Header(default="*/*")):
    if "*/*" in accept:
        return
    if not any(media in accept for media in SUPPORTED_MEDIA_TYPES):
        raise HTTPException(status_code=406, detail=f"Supported media types: {SUPPORTED_MEDIA_TYPES}")
