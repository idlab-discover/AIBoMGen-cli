import platform
import os
import time
import psutil
from celery import current_task
import docker
import subprocess
import json
from shared.minio_utils import download_file_from_minio, list_files_in_bucket, WORKER_SCANS_BUCKET


def extract_environment_details(task_logger, unique_dir, start_task_time, start_training_time, start_aibom_time):
    """
    Extract useful environment details for the training process, including GPU, Celery, Docker, and vulnerability information.

    Args:
        task_logger (Logger): Logger for logging task-specific information.
        unique_dir (str): Unique directory for the task.
        start_task_time (float): Task start time in seconds since epoch.
        start_training_time (float): Training start time in seconds since epoch.
        start_aibom_time (float): AIBoM generation start time in seconds since epoch.

    Returns:
        dict: A dictionary containing environment details.
    """
    try:
        task_logger.info("Extracting environment details...")

        # Extract details
        os_info = platform.system() + " " + platform.release()
        python_version = platform.python_version()
        tensorflow_version = get_tensorflow_version()
        cpu_count = psutil.cpu_count(logical=True)
        memory_total = psutil.virtual_memory().total // (1024 * 1024)  # in MB
        disk_usage = psutil.disk_usage('/').total // (1024 * 1024)  # in MB
        gpu_info = get_gpu_info(task_logger=task_logger)
        celery_task_info = get_celery_task_info(task_logger=task_logger)
        docker_info = get_docker_info(task_logger=task_logger)
        vulnerability_scan = fetch_latest_vulnerability_scan_from_minio(
            unique_dir, task_logger=task_logger)

        task_logger.info("Environment details extracted successfully.")

        return {
            "os": os_info,
            "python_version": python_version,
            "tensorflow_version": tensorflow_version,
            "cpu_count": cpu_count,
            "memory_total": memory_total,
            "disk_usage": disk_usage,
            "gpu_info": gpu_info,
            "celery_task_info": celery_task_info,
            "docker_info": docker_info,
            "vulnerability_scan": vulnerability_scan,
            "request_time": time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime(start_task_time)),
            "start_training_time": time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime(start_training_time)),
            "start_aibom_time": time.strftime("%Y-%m-%d %H:%M:%S", time.gmtime(start_aibom_time)),
            "training_time": start_aibom_time - start_training_time,
            "job_id": current_task.request.id if current_task else "Unknown",
            "unique_dir": unique_dir,
        }
    except Exception as e:
        task_logger.error(
            f"An error occurred while extracting environment details: {str(e)}")
        raise


def get_tensorflow_version(task_logger=None):
    """
    Get the installed TensorFlow version.

    Args:
        task_logger (Logger, optional): Logger for logging task-specific information.

    Returns:
        str: TensorFlow version or 'Unknown' if not installed.
    """
    try:
        if task_logger:
            task_logger.info("Retrieving TensorFlow version...")
        import tensorflow as tf
        version = tf.__version__
        if task_logger:
            task_logger.info(f"TensorFlow version retrieved: {version}")
        return version
    except ImportError:
        if task_logger:
            task_logger.warning("TensorFlow is not installed.")
        return "Unknown"
    except Exception as e:
        if task_logger:
            task_logger.error(f"Error retrieving TensorFlow version: {str(e)}")
        return "Unknown"


def get_gpu_info(task_logger=None):
    """
    Get GPU information using NVIDIA's NVML library.

    Args:
        task_logger (Logger, optional): Logger for logging task-specific information.

    Returns:
        list: A list of dictionaries containing GPU details. Returns an empty list when GPU is disabled or unavailable.
    """
    # Respect GPU toggle
    gpu_env = os.getenv("GPU", "true").strip().lower()
    use_gpu = gpu_env in ("1", "true", "yes", "on")

    if not use_gpu:
        if task_logger:
            task_logger.info(
                "GPU is disabled by configuration. Skipping GPU info retrieval.")
        return []

    try:
        # Import NVML lazily to avoid ImportError on CPU-only systems
        from pynvml import (
            nvmlInit,
            nvmlDeviceGetHandleByIndex,
            nvmlDeviceGetCount,
            nvmlDeviceGetName,
            nvmlDeviceGetMemoryInfo,
            nvmlShutdown,
        )

        if task_logger:
            task_logger.info("Retrieving GPU information...")
        nvmlInit()
        gpus = []
        device_count = nvmlDeviceGetCount()  # Get the number of GPUs
        for i in range(device_count):
            handle = nvmlDeviceGetHandleByIndex(i)
            name = nvmlDeviceGetName(handle).decode("utf-8")
            memory_info = nvmlDeviceGetMemoryInfo(handle)
            gpus.append({
                "name": name,
                "memory_total": memory_info.total // (1024 * 1024),  # in MB
                "memory_used": memory_info.used // (1024 * 1024),    # in MB
            })
        nvmlShutdown()
        if task_logger:
            task_logger.info("GPU information retrieved successfully.")
        return gpus
    except ImportError:
        if task_logger:
            task_logger.warning("pynvml not installed; skipping GPU info.")
        return []
    except Exception as e:
        if task_logger:
            task_logger.error(f"Error retrieving GPU info: {str(e)}")
        return []


def get_docker_info(task_logger=None):
    """
    Get Docker container and image information using the Docker SDK.

    Args:
        task_logger (Logger, optional): Logger for logging task-specific information.

    Returns:
        dict: Docker details or 'Not running in a Docker container' if not applicable.
    """
    try:
        if task_logger:
            task_logger.info(
                "Retrieving Docker container and image information...")
        # Check if running inside a Docker container
        if os.path.exists("/.dockerenv"):
            client = docker.from_env()
            container_id = os.getenv("HOSTNAME", "Unknown")
            container = client.containers.get(container_id)
            image = container.image
            docker_info = {
                "container_id": container_id,
                "image_name": image.tags[0] if image.tags else "Unknown",
                "image_id": image.id,
            }
            if task_logger:
                task_logger.info(
                    f"Docker information retrieved: {docker_info}")
            return docker_info
        else:
            if task_logger:
                task_logger.warning("Not running in a Docker container.")
            return "Not running in a Docker container"
    except docker.errors.DockerException as e:
        if task_logger:
            task_logger.error(f"Docker daemon is not accessible: {str(e)}")
        return f"Error retrieving Docker info: {str(e)}"
    except Exception as e:
        if task_logger:
            task_logger.error(f"Error retrieving Docker info: {str(e)}")
        return f"Error retrieving Docker info: {str(e)}"


def get_celery_task_info(task_logger=None):
    """
    Get information about the current Celery task.

    Args:
        task_logger (Logger, optional): Logger for logging task-specific information.

    Returns:
        dict: Celery task details or 'Unknown' if not running in a Celery task.
    """
    try:
        if task_logger:
            task_logger.info("Retrieving Celery task information...")
        if current_task:
            task_info = {
                "task_id": current_task.request.id,
                "task_name": current_task.name,
                "queue": current_task.request.delivery_info.get("routing_key", "Unknown"),
            }
            if task_logger:
                task_logger.info(
                    f"Celery task information retrieved: {task_info}")
            return task_info
        else:
            if task_logger:
                task_logger.warning("Not running in a Celery task.")
            return "Not running in a Celery task"
    except Exception as e:
        if task_logger:
            task_logger.error(f"Error retrieving Celery task info: {str(e)}")
        return f"Error retrieving Celery task info: {str(e)}"


def fetch_latest_vulnerability_scan_from_minio(unique_dir, task_logger=None):
    """
    Fetch the latest vulnerability scan results from MinIO.

    Args:
        unique_dir (str): Unique directory for the task.
        task_logger (Logger, optional): Logger for logging task-specific information.

    Returns:
        dict: The latest vulnerability scan summary.
    """
    try:
        if task_logger:
            task_logger.info(
                "Fetching the latest vulnerability scan results from MinIO...")

        # List all files in the vulnerability scans bucket
        bucket_prefix = "worker-vulnerability-scans/"
        files = list_files_in_bucket(bucket_prefix, WORKER_SCANS_BUCKET)

        if not files:
            if task_logger:
                task_logger.warning(
                    "No vulnerability scan files found in MinIO.")
            return {"error": "No vulnerability scan files found."}

        # Find the latest file based on timestamp in the filename
        latest_file = max(files, key=lambda x: x.split("_")
                          [-1].replace(".json", ""))
        local_file = f"/tmp/{os.path.basename(latest_file)}"

        # Download the latest file
        download_file_from_minio(latest_file, local_file, WORKER_SCANS_BUCKET)

        # Parse the JSON file
        with open(local_file, "r") as f:
            vulnerabilities = json.load(f)

        # Summarize vulnerabilities by severity
        summary = {}
        for result in vulnerabilities.get("Results", []):
            for vuln in result.get("Vulnerabilities", []):
                severity = vuln.get("Severity", "UNKNOWN")
                summary[severity] = summary.get(severity, 0) + 1

        if task_logger:
            task_logger.info(f"Latest vulnerability scan summary: {summary}")
        return summary

    except Exception as e:
        if task_logger:
            task_logger.error(
                f"Error fetching vulnerability scan results: {str(e)}")
        return {"error": f"Error fetching vulnerability scan results: {str(e)}"}
