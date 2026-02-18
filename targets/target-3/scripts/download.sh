#!/usr/bin/env bash
# download.sh – fetch models and weights from Hugging Face Hub
# Scanner should detect all huggingface-cli download and MODEL_* env assignments.

set -euo pipefail

# ── huggingface-cli download ──────────────────────────────────────────────────

# Bare model repo (detected: hf_cli_download)
huggingface-cli download meta-llama/Llama-3.1-8B

# With specific file pattern
huggingface-cli download bartowski/Meta-Llama-3.1-8B-Instruct-GGUF \
  --include "*.gguf" \
  --local-dir ./models/llama-gguf

# Quoted model ID
huggingface-cli download "mistralai/Mistral-7B-Instruct-v0.3"

# ── Environment variable assignments ─────────────────────────────────────────

# (detected: shell_model_env)
MODEL_NAME=google/gemma-2-9b-it
MODEL_ID=deepseek-ai/DeepSeek-R1-Distill-Qwen-7B
export HF_MODEL=stabilityai/stable-diffusion-xl-base-1.0
export HF_MODEL_ID=BAAI/bge-large-en-v1.5

# ── NOT detectable ────────────────────────────────────────────────────────────

# Model ID embedded in a URL – scanner won't match (URL prefix rejected)
MODEL_URL="https://huggingface.co/meta-llama/Llama-3.1-8B/resolve/main/config.json"
curl -L "$MODEL_URL" -o config.json

# Variable expansion – scanner won't follow variables
ORG="meta-llama"
VARIANT="Llama-3.1-8B"
huggingface-cli download "${ORG}/${VARIANT}"
