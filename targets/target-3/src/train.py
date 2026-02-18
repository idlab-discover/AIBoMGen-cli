"""
Training script exercising a wide range of HF loading patterns.

Detectable patterns (scanner should catch all of these):
  - from_pretrained: single/double quotes, multi-line open-paren split
  - from_pretrained: explicit pretrained_model_name_or_path= kwarg
  - pipeline(): positional and model= kwarg forms
  - PeftModel.from_pretrained: second positional arg (adapter path)
  - SentenceTransformer / CrossEncoder constructors
  - hf_hub_download: positional and repo_id= kwarg
  - snapshot_download: positional and repo_id= kwarg
  - HuggingFacePipeline.from_model_id: model_id= kwarg

NOT detectable (scanner limitations – variable indirection / dynamic strings):
  - Variable holding model ID passed to from_pretrained
  - f-string model IDs
  - os.environ model IDs
  - sys.argv model IDs
  - Conditionally assigned model ID string
"""

import os
import sys

from transformers import (
    AutoModelForCausalLM,
    AutoModelForSeq2SeqLM,
    AutoTokenizer,
    AutoConfig,
    pipeline,
    BitsAndBytesConfig,
)
from peft import PeftModel, PeftConfig
from huggingface_hub import hf_hub_download, snapshot_download
from sentence_transformers import SentenceTransformer, CrossEncoder
from langchain_community.llms import HuggingFacePipeline
import torch

# ── Standard from_pretrained ─────────────────────────────────────────────────

# Double-quote, single-line
model = AutoModelForCausalLM.from_pretrained("meta-llama/Llama-3.1-8B")
tokenizer = AutoTokenizer.from_pretrained("meta-llama/Llama-3.1-8B")

# Single-quote variant
config = AutoConfig.from_pretrained('mistralai/Mistral-7B-v0.1')

# Multi-line open-paren split (scanner stitches lines ending with '(')
quant_model = AutoModelForCausalLM.from_pretrained(
    "Qwen/Qwen2.5-7B-Instruct",
    torch_dtype=torch.bfloat16,
    device_map="auto",
)

# Explicit pretrained_model_name_or_path= keyword argument
seq2seq = AutoModelForSeq2SeqLM.from_pretrained(
    pretrained_model_name_or_path="google/flan-t5-xxl",
    device_map="auto",
)

# ── pipeline() ───────────────────────────────────────────────────────────────

# Positional: pipeline(<task>, <model-id>)
pipe_pos = pipeline("text-generation", "facebook/opt-1.3b")

# Named model= kwarg
pipe_kwarg = pipeline(
    "zero-shot-classification",
    model="facebook/bart-large-mnli",
    device=0,
)

# ── PEFT / LoRA ──────────────────────────────────────────────────────────────

# PeftModel.from_pretrained(base_model_variable, "adapter-repo/name")
# First arg is a loaded model object (not detectable), second is detectble adapter ID.
base = AutoModelForCausalLM.from_pretrained("meta-llama/Llama-3.1-8B-Instruct")
peft_model = PeftModel.from_pretrained(base, "timdettmers/guanaco-33b-merged")
# Note: the adapter ID above is the SECOND positional arg – not matched by the
# standard from_pretrained rule (known scanner limitation). Make it scannable
# by also referencing it via hf_hub_download:
_ = hf_hub_download(repo_id="timdettmers/guanaco-33b-merged",
                    filename="adapter_config.json")

# ── SentenceTransformers ──────────────────────────────────────────────────────

embedder = SentenceTransformer("BAAI/bge-large-en-v1.5")
reranker = CrossEncoder("cross-encoder/ms-marco-MiniLM-L-6-v2")

# ── hf_hub_download ───────────────────────────────────────────────────────────

# Positional
gguf_path = hf_hub_download(
    "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF", "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf")

# Keyword repo_id=
config_path = hf_hub_download(
    repo_id="google/gemma-2-9b-it",
    filename="config.json",
)

# ── snapshot_download ─────────────────────────────────────────────────────────

# Positional
snap1 = snapshot_download("deepseek-ai/DeepSeek-R1-Distill-Qwen-7B")

# Keyword repo_id=
snap2 = snapshot_download(
    repo_id="microsoft/phi-4",
    ignore_patterns=["*.pt", "*.bin"],
)

# ── HuggingFacePipeline (LangChain) ──────────────────────────────────────────

lc_pipe = HuggingFacePipeline.from_model_id(
    model_id="tiiuae/falcon-7b-instruct",
    task="text-generation",
    pipeline_kwargs={"max_new_tokens": 100},
)

# ── NOT detectable: variable indirection ─────────────────────────────────────

# string visible but not in a call
MODEL_ID = "meta-llama/Llama-3.1-70B-Instruct"
model_indirect = AutoModelForCausalLM.from_pretrained(
    MODEL_ID)  # variable → not caught

# NOT detectable: f-string
variant = "8B"
model_fstring = AutoModelForCausalLM.from_pretrained(
    f"meta-llama/Llama-3.1-{variant}")

# NOT detectable: env variable
model_env = AutoModelForCausalLM.from_pretrained(
    os.environ.get("MODEL_ID", ""))

# NOT detectable: local path (correctly ignored)
model_local1 = AutoModelForCausalLM.from_pretrained(
    "./checkpoints/my-finetuned-model")
model_local2 = AutoModelForCausalLM.from_pretrained("/mnt/models/llama-8b")
model_local3 = AutoModelForCausalLM.from_pretrained(
    "../weights/checkpoint-500")

# NOT detectable: generic single-segment local dir name passed to from_pretrained
# (single-segment IS caught by from_pretrained rule – this is by design)
# Known false-positive vector: AutoModelForCausalLM.from_pretrained(<local-dir>)
# Use full absolute or relative paths to avoid false positives.
