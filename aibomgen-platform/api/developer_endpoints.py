import os
import shutil
import uuid
from typing import Literal, Optional
from fastapi.responses import RedirectResponse
import yaml
from celery.result import AsyncResult
from database import SessionLocal
from fastapi import (APIRouter, Depends, File, Form, HTTPException, Request,
                     UploadFile)
from fastapi_azure_auth.user import User
from models import Job
from shared.minio_utils import (TRAINING_BUCKET, generate_presigned_url,
                                list_files_in_bucket, upload_file_to_minio)
from shared.zip_utils import ZipValidationError, validate_zip_file
from slowapi import Limiter
from slowapi.util import get_remote_address
from sqlalchemy.orm import Session
from auth_utils import get_current_user
from celery_config import celery_app

# === Router Setup ===
developer_router = APIRouter(prefix="/developer", tags=["Developer Endpoints"])

# Limiter setup
limiter = Limiter(key_func=get_remote_address)
ENABLE_RATE_LIMIT = os.getenv("ENABLE_RATE_LIMIT", "false").lower() == "true"

# === Database Dependency ===


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# === Developer Endpoints ===


def _submit_job_decorator(func):
    if ENABLE_RATE_LIMIT:
        return developer_router.post("/submit_job_by_model_and_data", dependencies=[Depends(get_current_user)])(limiter.limit("5/minute")(func))
    else:
        return developer_router.post("/submit_job_by_model_and_data", dependencies=[Depends(get_current_user)])(func)


@_submit_job_decorator
async def submit_job(
    request: Request,
    # Use Depends to get the user object
    user: User = Depends(get_current_user),
    db: Session = Depends(get_db),  # Inject the database session
    # File uploads
    model: UploadFile = File(
        ..., description="Model file to be trained (currently only .keras for tensorflow framework)."),
    dataset: UploadFile = File(
        ..., description="Dataset file for training (currently only csv and .zip for image data)."),
    dataset_definition: UploadFile = File(
        ..., description="Dataset definition file (currently in YAML, see spec in project README)."),

    # Model metadata (moved to Form)
    framework: Literal["TensorFlow 2.16.1"] = Form(
        ..., description="Currently, only TensorFlow 2.16.1 is supported."),
    model_name: Optional[str] = Form("", description="Name of the model."),
    model_version: Optional[str] = Form(
        "", description="Version of the model."),
    model_description: Optional[str] = Form(
        "", description="Description of the model."),
    author: Optional[str] = Form("", description="Author of the model."),
    model_type: Optional[str] = Form(
        "", description="Type of the model (e.g., Image Classification)."),
    base_model: Optional[str] = Form(
        "", description="Base model used (e.g., ResNet50)."),
    base_model_source: Optional[str] = Form(
        "", description="Source URL of the base model."),
    intended_use: Optional[str] = Form(
        "", description="Intended use of the model."),
    out_of_scope: Optional[str] = Form(
        "", description="Out-of-scope use cases."),
    misuse_or_malicious: Optional[str] = Form(
        "", description="Misuse or malicious use cases."),
    license_name: Optional[str] = Form(
        "", description="License name for the model."),

    # Training parameters for model.fit (moved to Form)
    epochs: Optional[int] = Form(50, description="Number of epochs to train."),
    validation_split: Optional[float] = Form(
        0.2, description="Fraction of data to use for validation."),
    initial_epoch: Optional[int] = Form(
        0, description="Epoch at which to start training."),
    batch_size: Optional[int] = Form(
        32, description="Size of the batches of data. Default is 32."),
    steps_per_epoch: Optional[int] = Form(
        None, description="Total number of steps per epoch. If None or zero, it will be calculated with batch size."),
    validation_steps: Optional[int] = Form(
        None, description="Number of steps for validation. If None or zero, it will be calculated with batch size."),
    validation_freq: Optional[int] = Form(
        1, description="Specifies how many training epochs to run before a new validation run is performed. Default is 1 (every epoch)."),
):
    """
Submit a training job with model and dataset files, along with optional metadata and training parameters.

Args:

    Files Uploads:
        - model: The model file to be trained (currently only .keras for tensorflow framework).
        - dataset: The dataset file for training (currently only csv and .zip for image data). 
        - dataset_definition: The dataset definition file (currently in YAML, see spec in project README).

    Required Metadata:
        - framework: Framework used (currently only TensorFlow 2.16.1 is supported).

    Model Metadata:
        - model_name: Name of the model.
        - model_version: Version of the model.
        - model_description: Description of the model.
        - author: Author of the model.
        - model_type: Type of the model (e.g., Image Classification).
        - base_model: Base model used (e.g., ResNet50).
        - base_model_source: Source URL of the base model.
        - intended_use: Intended use of the model.
        - out_of_scope: Out-of-scope use cases.
        - misuse_or_malicious: Misuse or malicious use cases.
        - license_name: License name for the model.

    Training Parameters (optional):
        - epochs: Number of epochs to train.
        - validation_split: Fraction of data to use for validation.
        - initial_epoch: Epoch at which to start training.
        - batch_size: Size of the batches of data.
        - steps_per_epoch: Total number of steps per epoch.
        - validation_steps: Number of steps for validation.
        - validation_freq: Frequency of validation runs.


"""
    # Save the user ID from the Azure token
    user_id = user.claims.get("oid")  # Get the user's Azure Object ID

    try:
        # Generate a unique directory for the job
        unique_dir = str(uuid.uuid4())

        # Save files temporarily
        temp_dir = os.path.join("/tmp", unique_dir)
        os.makedirs(temp_dir, exist_ok=True)

        model_path = os.path.join(temp_dir, model.filename)
        dataset_path = os.path.join(temp_dir, dataset.filename)
        dataset_definition_path = os.path.join(
            temp_dir, dataset_definition.filename)

        with open(model_path, "wb") as buffer:
            shutil.copyfileobj(model.file, buffer)
        with open(dataset_path, "wb") as buffer:
            shutil.copyfileobj(dataset.file, buffer)
        with open(dataset_definition_path, "wb") as buffer:
            shutil.copyfileobj(dataset_definition.file, buffer)

        # Load the dataset definition to determine the dataset type
        with open(dataset_definition_path, "r") as f:
            dataset_definition_yaml = yaml.safe_load(f)

        dataset_type = dataset_definition_yaml.get(
            "type", "csv")  # Default to 'csv' if not specified

        # Validate the dataset .zip file only if the type is 'image'
        if dataset_type == "image":
            try:
                validate_zip_file(dataset_path)  # Ensure the .zip file is safe
            except ZipValidationError as e:
                raise HTTPException(status_code=400, detail=str(e))

        # Upload files to MinIO
        model_url = upload_file_to_minio(
            model_path, f"{unique_dir}/model/{model.filename}", TRAINING_BUCKET)
        dataset_url = upload_file_to_minio(
            dataset_path, f"{unique_dir}/dataset/{dataset.filename}", TRAINING_BUCKET)
        dataset_definition_url = upload_file_to_minio(
            dataset_definition_path, f"{unique_dir}/definition/{dataset_definition.filename}", TRAINING_BUCKET)

        # Send Celery task with file URLs
        task = celery_app.send_task(
            'tasks.run_training',
            args=[
                unique_dir, model_url, dataset_url, dataset_definition_url,
                {
                    "model_name": model_name,
                    "model_version": model_version,
                    "model_description": model_description,
                    "author": author,
                    "framework": framework,
                    "model_type": model_type,
                    "base_model": base_model,
                    "base_model_source": base_model_source,
                    "intended_use": intended_use,
                    "out_of_scope": out_of_scope,
                    "misuse_or_malicious": misuse_or_malicious,
                    "license_name": license_name,
                },
                {
                    "epochs": epochs,
                    "validation_split": validation_split,
                    "initial_epoch": initial_epoch,
                    "batch_size": batch_size,
                    "steps_per_epoch": steps_per_epoch,
                    "validation_steps": validation_steps,
                    "validation_freq": validation_freq,
                }
            ],
            queue='training_queue'
        )

        # Store job metadata in the database
        job = Job(
            id=task.id,
            user_id=user_id,
            unique_dir=unique_dir,
        )
        db.add(job)
        db.commit()

        return {"job_id": task.id, "status": "Training started", "unique_dir": unique_dir}

    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Task submission failed: {str(e)}")


@developer_router.get("/job_status/{job_id}", dependencies=[Depends(get_current_user)])
async def job_status(job_id: str, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    user_id = user.claims.get("oid")

    job = db.query(Job).filter(Job.id == job_id).first()
    if not job:
        raise HTTPException(status_code=404, detail="Job not found.")
    if job.user_id != user_id:
        raise HTTPException(
            status_code=403, detail="You do not have permission to access this job.")

    celery_task = AsyncResult(job_id, app=celery_app)
    if not celery_task:
        raise HTTPException(status_code=404, detail="Job not found.")
    return {"status": celery_task.status, "result": celery_task.result}


@developer_router.get("/job_artifacts/{job_id}", dependencies=[Depends(get_current_user)])
async def get_job_artifacts(job_id: str, user: User = Depends(get_current_user), db: Session = Depends(get_db)):
    user_id = user.claims.get("oid")

    job = db.query(Job).filter(Job.id == job_id).first()
    if not job:
        raise HTTPException(status_code=404, detail="Job not found.")
    if job.user_id != user_id:
        raise HTTPException(
            status_code=403, detail="You are not authorized to access this job.")

    artifacts = list_files_in_bucket(f"{job.unique_dir}/", TRAINING_BUCKET)
    if not artifacts:
        raise HTTPException(
            status_code=404, detail="No artifacts found for this job.")
    return {"job_id": job_id, "artifacts": artifacts}


@developer_router.get("/job_artifacts/{job_id}/{artifact_name}", dependencies=[Depends(get_current_user)])
async def download_artifact(job_id: str, artifact_name: str, user: User = Depends(get_current_user), db: Session = Depends(get_db), redirect: bool = True, test_mode: bool = True):
    user_id = user.claims.get("oid")

    job = db.query(Job).filter(Job.id == job_id).first()
    if not job:
        raise HTTPException(status_code=404, detail="Job not found.")
    if job.user_id != user_id:
        raise HTTPException(
            status_code=403, detail="You are not authorized to access this job.")

    unique_dir = job.unique_dir
    all_files = list_files_in_bucket(f"{unique_dir}/", TRAINING_BUCKET)
    if not all_files:
        raise HTTPException(
            status_code=404, detail="No files found for this job.")

    matching_files = [
        file for file in all_files if file.endswith(f"/{artifact_name}")]
    if not matching_files:
        raise HTTPException(
            status_code=404, detail=f"Artifact '{artifact_name}' not found.")
    if len(matching_files) > 1:
        raise HTTPException(
            status_code=400, detail=f"Multiple artifacts named '{artifact_name}' found.")

    object_name = matching_files[0]
    presigned_url = generate_presigned_url(
        object_name, TRAINING_BUCKET, expiration=3600)  # URL valid for 1 hour

    if test_mode:
        presigned_url = presigned_url.replace("minio:9000", "localhost:9000")

    if redirect:
        return RedirectResponse(url=presigned_url)
    else:
        return {"artifact_name": artifact_name, "url": presigned_url}
