FROM tensorflow/tensorflow:2.16.1-gpu

WORKDIR /app

# Install Docker CLI and dependencies
RUN apt-get update && apt-get install -y docker.io curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set permissions for /app directory
RUN chmod 755 /app

# Copy requirements and install dependencies
COPY ./worker/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the shared directory
COPY ./shared /app/shared

# Copy the worker code
COPY ./worker /app 

# Ensure entrypoint script has correct line endings and is executable
RUN sed -i 's/\r$//' /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Run as root (default user in Docker)
ENTRYPOINT ["/app/entrypoint.sh"]