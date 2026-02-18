FROM python:3.11-slim

# ENV assignments (detected: shell_model_env)
ENV MODEL_NAME=meta-llama/Llama-3.1-8B-Instruct
ENV HF_MODEL_ID=BAAI/bge-large-en-v1.5
ENV HF_MODEL=mistralai/Mistral-7B-Instruct-v0.3

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# huggingface-cli download inside RUN (detected: hf_cli_download)
RUN huggingface-cli download google/gemma-2-9b-it --local-dir /models/gemma

# ARG then ENV – scanner matches the ENV line
ARG HF_TOKEN
ENV HF_TOKEN=${HF_TOKEN}

COPY . .

# NOT detectable: model ID constructed from ARGs at build time
ARG MODEL_ORG=deepseek-ai
ARG MODEL_NAME_ARG=DeepSeek-R1-Distill-Qwen-7B
RUN echo "Loading ${MODEL_ORG}/${MODEL_NAME_ARG}"

CMD ["python", "src/train.py"]
