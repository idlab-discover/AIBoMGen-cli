import os
import webbrowser
from dotenv import load_dotenv

# Does not work in WSL!!

# Load environment variables from .env file
dotenv_path = os.path.join(os.path.dirname(__file__), '../.env')
load_dotenv(dotenv_path)

# Get credentials from environment variables
rabbitmq_user = os.getenv("RABBITMQ_USER")
rabbitmq_password = os.getenv("RABBITMQ_PASSWORD")
minio_user = os.getenv("MINIO_ROOT_USER")
minio_password = os.getenv("MINIO_ROOT_PASSWORD")
flower_auth = os.getenv("FLOWER_BASIC_AUTH")

# URLs with credentials
urls = {
    "FastAPI Docs": "http://localhost:8000/docs",
    "RabbitMQ Management UI": f"http://localhost:15672",
    "MinIO Console": f"http://localhost:9001",
    "Flower Dashboard": f"http://{flower_auth}@localhost:5555",
}

# Open each URL in the default web browser
for name, url in urls.items():
    print(f"Opening {name}: {url}")
    webbrowser.open(url)

print("Dashboards opened in your default browser.")
