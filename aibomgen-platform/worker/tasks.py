from celery_config import celery_app
import os
import tensorflow as tf
import time
import yaml
import pandas as pd
import json
from transform_to_cyclonedx import serialize_bom, sign_and_include_bom_as_property, transform_to_cyclonedx, sign_bom
from bom_data_generator import generate_basic_bom_data
from shared.minio_utils import upload_file_to_minio, download_file_from_minio, TRAINING_BUCKET, remove_file_from_minio
from shared.zip_utils import ZipValidationError, validate_and_extract_zip
import logging
from in_toto_link_generator import generate_in_toto_link
from shared.in_toto_utils import load_signer, record_artifact_as_dict
from environment_extractor import extract_environment_details

from training_logic import (
    load_csv_dataset_with_definition,
    load_image_dataset,
    load_TFRecordDataset_with_definition,
    validate_model_and_dataset_definition,
    apply_preprocessing
)


@celery_app.task(name="tasks.run_training", time_limit=3600)
def run_training(unique_dir, model_url, dataset_url, dataset_definition_url, optional_params=None, fit_params=None):
    """Training task with support for tabular and image data."""

    e = None

    # Create a temporary directory
    temp_dir = os.path.join("/tmp", unique_dir)
    os.makedirs(temp_dir, exist_ok=True)

    # Define a unique log file for this task
    logs_path = os.path.join(temp_dir, "logs.log")

    # Create a logger for this task
    task_logger = logging.getLogger(f"task_logger_{unique_dir}")
    task_logger.setLevel(logging.INFO)

    # File handler for writing logs to a file
    file_handler = logging.FileHandler(logs_path)
    file_handler.setLevel(logging.INFO)
    file_formatter = logging.Formatter(
        "%(asctime)s - %(levelname)s - %(message)s")
    file_handler.setFormatter(file_formatter)

    # Console handler for displaying logs in the terminal
    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.INFO)
    console_formatter = logging.Formatter(
        "%(asctime)s - %(levelname)s - %(message)s")
    console_handler.setFormatter(console_formatter)

    # Add handlers to the logger
    task_logger.addHandler(file_handler)
    task_logger.addHandler(console_handler)

    # Avoid duplicate logs by disabling propagation to the root logger
    task_logger.propagate = False

    task_logger.info("Starting training task...")
    task_logger.info("Logging system initialized successfully.")

    try:

        # Determine GPU usage from environment
        gpu_env = os.getenv("GPU", "true").strip().lower()
        use_gpu = gpu_env in ("1", "true", "yes", "on")

        cpus = tf.config.list_physical_devices('CPU')
        if not cpus:
            task_logger.warning("No CPU devices found!")
        else:
            task_logger.info(f"CPUs available: {[cpu.name for cpu in cpus]}")

        if use_gpu:
            # Confirm GPU availability
            gpus = tf.config.list_physical_devices('GPU')
            if not gpus:
                task_logger.warning(
                    "No GPU devices found! Falling back to CPU.")
            else:
                task_logger.info(
                    f"GPUs available: {[gpu.name for gpu in gpus]}")

            # Device selection
            if len(gpus) > 0 and len(cpus) > 0:
                task_logger.info(
                    "GPU is enabled and available. Using GPU for training.")
                tf.config.set_visible_devices(gpus[0], 'GPU')
            elif len(cpus) > 0:
                task_logger.info(
                    "GPU enabled but not available; using CPU for training.")
                tf.config.set_visible_devices(cpus[0], 'CPU')
            else:
                raise RuntimeError("No available devices for training.")
        else:
            task_logger.info(
                "GPU usage is disabled via configuration. Forcing CPU.")
            if len(cpus) > 0:
                tf.config.set_visible_devices(cpus[0], 'CPU')
            else:
                raise RuntimeError("No available CPU devices for training.")

        start_task_time = time.time()
        start_task_time_utc = time.strftime(
            "%Y-%m-%d %H:%M:%S", time.gmtime(start_task_time))
        task_logger.info(f"Task started at UTC: {start_task_time_utc}")

        # Define paths for downloaded files
        model_dir = os.path.join(temp_dir, "model")
        dataset_dir = os.path.join(temp_dir, "dataset")
        dataset_definition_dir = os.path.join(temp_dir, "definition")

        # Ensure directories exist
        os.makedirs(model_dir, exist_ok=True)
        os.makedirs(dataset_dir, exist_ok=True)
        os.makedirs(dataset_definition_dir, exist_ok=True)

        # Define filenames and paths
        model_filename = model_url.split("/")[-1]
        dataset_filename = dataset_url.split("/")[-1]
        dataset_definition_filename = dataset_definition_url.split("/")[-1]

        model_path = os.path.join(model_dir, model_filename)
        dataset_path = os.path.join(dataset_dir, dataset_filename)
        dataset_definition_path = os.path.join(
            dataset_definition_dir, dataset_definition_filename)

        # Download files from MinIO
        task_logger.info("Downloading files from MinIO...")
        download_file_from_minio(
            f"{unique_dir}/model/{model_filename}", model_path, TRAINING_BUCKET)
        download_file_from_minio(
            f"{unique_dir}/dataset/{dataset_filename}", dataset_path, TRAINING_BUCKET)
        download_file_from_minio(
            f"{unique_dir}/definition/{dataset_definition_filename}", dataset_definition_path, TRAINING_BUCKET)

        # Load dataset definition
        task_logger.info("Loading dataset definition...")
        with open(dataset_definition_path, "r") as f:
            dataset_definition = yaml.safe_load(f)

        # Load dataset based on type
        dataset_type = dataset_definition.get("type", "csv")
        task_logger.info(f"Dataset type: {dataset_type}")
        if dataset_type == "csv":
            dataset = load_csv_dataset_with_definition(
                task_logger, dataset_path, dataset_definition)
        elif dataset_type == "image":
            # Validate and extract the dataset .zip file
            dataset_zip_path = dataset_path
            if not os.path.exists(dataset_zip_path):
                raise FileNotFoundError(
                    f"Dataset file {dataset_zip_path} does not exist.")
            dataset_extracted_path = os.path.join(temp_dir, "dataset_unzipped")
            os.makedirs(dataset_extracted_path, exist_ok=True)
            try:
                task_logger.info(
                    "Validating and extracting dataset zip file...")
                # Validate and extract the zip file
                validate_and_extract_zip(
                    dataset_zip_path, dataset_extracted_path)
                task_logger.info("Dataset zip file extracted successfully.")
                # Load the dataset from the extracted path
                dataset = load_image_dataset(
                    task_logger, dataset_extracted_path, dataset_definition)

            except ZipValidationError as e:
                raise Exception(f"Dataset validation failed: {str(e)}")

        elif dataset_type == "tfrecord":
            dataset = load_TFRecordDataset_with_definition(
                task_logger, dataset_path, dataset_definition)
        else:
            raise ValueError(f"Unsupported dataset type: {dataset_type}")

        # Load model
        task_logger.info("Loading model...")
        model = tf.keras.models.load_model(model_path)

        # Validate compatibility
        task_logger.info("Validating model and dataset compatibility...")
        validate_model_and_dataset_definition(model, dataset_definition)

        # Training start
        start_training_time = time.time()
        start_training_time_utc = time.strftime(
            "%Y-%m-%d %H:%M:%S", time.gmtime(start_training_time))
        task_logger.info(f"Training started at UTC: {start_training_time_utc}")

        # Split the dataset into training and validation subsets
        validation_split = fit_params.get(
            "validation_split", 0.2)  # Default to 20% validation
        if isinstance(dataset, tf.data.Dataset) and validation_split > 0.0:
            task_logger.info(
                "Splitting dataset into training and validation subsets...")
            # Calculate the number of samples in the dataset
            dataset_size = sum(1 for _ in dataset)
            val_size = int(validation_split * dataset_size)
            train_size = dataset_size - val_size

            # Split the dataset
            train_dataset = dataset.take(train_size)
            val_dataset = dataset.skip(train_size)
        else:
            train_dataset = dataset
            val_dataset = None

        # Train the model using fit_params
        task_logger.info("Starting model training...")
        model.fit(
            x=train_dataset,
            validation_data=val_dataset,  # Use the validation dataset if available
            epochs=fit_params.get("epochs", 50),
            verbose=2,  # 2 for one line per epoch
            initial_epoch=fit_params.get("initial_epoch", 0),
            steps_per_epoch=fit_params.get(
                "steps_per_epoch", train_size // fit_params.get("batch_size", 32)),
            validation_steps=fit_params.get(
                "validation_steps", val_size // fit_params.get("batch_size", 32)) if val_dataset else None,
            validation_freq=fit_params.get("validation_freq", 1),
        )
        task_logger.info("Model training completed.")

        # Define paths for output artifacts
        trained_model_path = os.path.join(temp_dir, "trained_model.keras")
        metrics_path = os.path.join(temp_dir, "metrics.json")
        bom_path = os.path.join(temp_dir, "cyclonedx_bom.json")

        # Save the trained model
        task_logger.info("Saving trained model...")
        model.save(trained_model_path)

        # Save training metrics
        task_logger.info("Saving training metrics...")
        with open(metrics_path, "w") as f:
            json.dump(model.history.history, f)

        upload_file_to_minio(
            trained_model_path, f"{unique_dir}/output/trained_model.keras", TRAINING_BUCKET)
        upload_file_to_minio(
            metrics_path, f"{unique_dir}/output/metrics.json", TRAINING_BUCKET)

        # in-toto LINK ----------------------------------------------------------------------

        # start AIBoM generation time
        start_aibom_time = time.time()
        start_aibom_time_utc = time.strftime(
            "%Y-%m-%d %H:%M:%S", time.gmtime(start_aibom_time))
        task_logger.info(
            f"AIBoM generation started at UTC: {start_aibom_time_utc}")
        task_logger.info("Generating BOM data...")

        # Load the persistent private key and public key from /run/secrets
        private_key_path = "/run/secrets/worker_private_key"
        public_key_path = "/run/secrets/worker_public_key"
        worker_signer = load_signer(private_key_path, public_key_path)

        # Record input and output artifacts for in-toto
        in_toto_materials = {
            f"{unique_dir}/model/{model_filename}": record_artifact_as_dict(model_path),
            f"{unique_dir}/dataset/{dataset_filename}": record_artifact_as_dict(dataset_path),
            f"{unique_dir}/definition/{dataset_definition_filename}": record_artifact_as_dict(dataset_definition_path),
        }

        in_toto_products = {
            f"{unique_dir}/output/trained_model.keras": record_artifact_as_dict(trained_model_path),
            f"{unique_dir}/output/metrics.json": record_artifact_as_dict(metrics_path),
        }

        # Record input and output artifacts with local paths for BOM generation
        materials = {
            f"{unique_dir}/model/{model_filename}": {
                "sha256": record_artifact_as_dict(model_path)["sha256"],
                "local_path": model_path,  # Pass the local path directly
            },
            f"{unique_dir}/dataset/{dataset_filename}": {
                "sha256": record_artifact_as_dict(dataset_path)["sha256"],
                "local_path": dataset_path,  # Pass the local path directly
            },
            f"{unique_dir}/definition/{dataset_definition_filename}": {
                "sha256": record_artifact_as_dict(dataset_definition_path)["sha256"],
                "local_path": dataset_definition_path,  # Pass the local path directly
            },
        }

        products = {
            f"{unique_dir}/output/trained_model.keras": {
                "sha256": record_artifact_as_dict(trained_model_path)["sha256"],
                "local_path": trained_model_path,  # Pass the local path directly
            },
            f"{unique_dir}/output/metrics.json": {
                "sha256": record_artifact_as_dict(metrics_path)["sha256"],
                "local_path": metrics_path,  # Pass the local path directly
            },
        }

        # Generate the in-toto link file
        link_file_path = generate_in_toto_link(
            task_name="run_training",
            materials=in_toto_materials,
            products=in_toto_products,
            command=["python", "tasks.py", "run_training"],
            signer=worker_signer,
            temp_dir=temp_dir,
            task_logger=task_logger,
        )

        # Upload the in-toto link file to MinIO
        task_logger.info("Uploading in-toto link file to MinIO...")
        # Ensure the file in minio also has the keyid in the name by using the basename of the link file
        link_file_minio_path = f"{unique_dir}/output/{os.path.basename(link_file_path)}"
        upload_file_to_minio(
            link_file_path, link_file_minio_path, TRAINING_BUCKET)
        task_logger.info("in-toto link file uploaded successfully.")

        # AIBOM -------------------------------------------------------------------------

        # Generate the BOM
        task_logger.info("Generating BOM...")

        environment_details = extract_environment_details(
            task_logger=task_logger,
            start_task_time=start_task_time,
            start_training_time=start_training_time,
            start_aibom_time=start_aibom_time,
            unique_dir=unique_dir,
        )

        # Generate BOM data
        bom_data = generate_basic_bom_data(
            task_logger=task_logger,
            environment=environment_details,
            materials=materials,
            products=products,
            fit_params=fit_params,
            optional_params=optional_params,
            link_file_minio_path=link_file_minio_path,
            unique_dir=unique_dir,
        )

        # Transform to CycloneDX format
        task_logger.info("Transforming BOM data to CycloneDX format...")

        cyclonedx_bom = transform_to_cyclonedx(bom_data)
        task_logger.info(f"Signing BOM data...")
        sign_and_include_bom_as_property(cyclonedx_bom, private_key_path)
        task_logger.info(f"BOM signed")

        # Verify the signature (currently Signature not supported in CycloneDX so its added as a property in the metadata)
        if not any(prop.name == "BOM Signature" for prop in cyclonedx_bom.metadata.properties):
            raise RuntimeError("BOM signature was not added successfully.")

        task_logger.info(f"serializing BOM data...")
        serialize_bom(cyclonedx_bom, bom_path)
        task_logger.info(f"BOM serialized: {bom_path}")

        # Upload output artifacts to MinIO
        task_logger.info("Uploading bom to MinIO...")
        upload_file_to_minio(
            bom_path, f"{unique_dir}/output/cyclonedx_bom.json", TRAINING_BUCKET)

        task_logger.info("Task completed successfully.")
        result = {
            "training_status": "training job completed",
            "unique_dir": unique_dir,
            "job_id": celery_app.current_task.request.id,
            "message": "Training completed successfully and AIBoM generated.",
        }
        return result
    except Exception as ex:
        e = ex  # Assign the exception to the variable
        task_logger.error(f"An error occurred: {str(e)}")

        # remove from minio if the task fails
        try:
            task_logger.info(
                f"Removing output files from MinIO if they exist for unique_dir: {unique_dir}")
            remove_file_from_minio(
                f"{unique_dir}/output/trained_model.keras", TRAINING_BUCKET)
            remove_file_from_minio(
                f"{unique_dir}/output/metrics.json", TRAINING_BUCKET)
            remove_file_from_minio(
                f"{unique_dir}/output/cyclonedx_bom.json", TRAINING_BUCKET)
            remove_file_from_minio(
                f"{unique_dir}/output/{os.path.basename(link_file_path)}", TRAINING_BUCKET)
        except FileNotFoundError:
            task_logger.warning(
                f"Output files not found in MinIO for unique_dir: {unique_dir}")
        except Exception as ex:
            task_logger.error(f"Failed to remove files from MinIO: {str(e)}")

        return {
            "training_status": "training job failed",
            "unique_dir": unique_dir,
            "error": str(e),
        }
    finally:
        task_logger.info(
            f"Task {celery_app.current_task.request.id} completed.")

        # Verify the logs file before uploading
        if os.path.exists(logs_path):
            file_size = os.path.getsize(logs_path)
            if file_size > 0:
                task_logger.info(
                    f"Uploading application_logs to MinIO. Size: {file_size} bytes")
                upload_file_to_minio(
                    logs_path, f"{unique_dir}/output/logs.log", TRAINING_BUCKET)
            else:
                task_logger.error("logs.log is empty. Skipping upload.")
        else:
            task_logger.error("logs.log does not exist. Skipping upload.")

        # Remove handlers to avoid memory leaks
        task_logger.removeHandler(file_handler)
        task_logger.removeHandler(console_handler)


# model.fit(
#     x=None, # data (could be: numpy array, tensor, dict mapping, tf.data.Dataset, keras.utils.PyDataset)
#       y=None, # label (could be: numpy array, (if x is dataset,generator or pydataset, y should not be specified))
#       batch_size=None, # Number of samples per gradient update, default 32 (if x is dataset,generator or pydataset, batch_size should not be specified)
#     epochs=1, # Number of epochs to train
#     verbose='auto', # 0: silent, 1: progress bar, 2: one line per epoch (when writing to file)
#       callbacks=None, # (tensorboard)
#       validation_split=0.0, # Fraction of train data used to validate (if x is dataset,generator or pydataset, validation_split should not be specified)
#     validation_data=None, # data to val (could be: numpy array, tensor, dict mapping, tf.data.Dataset, keras.utils.PyDataset)
#       shuffle=True, # shuffle x after each epoch (if x is dataset,generator or pydataset, shuffle should not be specified)
#     class_weight=None, # Optional dictionary mapping class indices (integers) to a weight (float) value, used for weighting the loss function (during training only). This can be useful to tell the model to "pay more attention" to samples from an under-represented class.
#       sample_weight=None, # Optional NumPy array of weights for the training samples, used for weighting the loss function (during training only). Y
#     initial_epoch=0, # Epoch at which to start training (useful for resuming a previous training run).
#     steps_per_epoch=None, # Total number of steps (batches of samples) before declaring one epoch finished and starting the next epoch. Epoch at which to start training (useful for resuming a previous training run).
#     validation_steps=None, # Only relevant if validation_data is provided. Total number of steps (batches of samples) to draw before stopping when performing validation at the end of every epoch.
#     validation_batch_size=None, # Number of samples per validation batch.
#     validation_freq=1 # Specifies how many training epochs to run before a new validation run is performed
# )
