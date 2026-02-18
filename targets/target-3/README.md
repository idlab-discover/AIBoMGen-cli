---
language:
  - en
  - ar
license: apache-2.0
base_model: meta-llama/Llama-3.1-8B
model_id: my-org/llama-3.1-8b-arabic-sft
tags:
  - text-generation
  - peft
  - lora
---

# my-org/llama-3.1-8b-arabic-sft

A LoRA fine-tune of **meta-llama/Llama-3.1-8B** on Arabic instruction data.

## Usage

```python
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import PeftModel

base = AutoModelForCausalLM.from_pretrained("meta-llama/Llama-3.1-8B")
model = PeftModel.from_pretrained(base, "my-org/llama-3.1-8b-arabic-sft")
tokenizer = AutoTokenizer.from_pretrained("meta-llama/Llama-3.1-8B")
```

## Evaluation

We compared against the following baselines (detected via inline markdown body scan):

- `mistralai/Mistral-7B-Instruct-v0.3` — baseline 1
- `Qwen/Qwen2.5-7B-Instruct` — baseline 2
- `google/gemma-2-9b-it` — baseline 3

The adapter was trained using `BAAI/bge-large-en-v1.5` embeddings for retrieval.

## Limitations

Single-segment names like `bert-base-uncased` appearing in the README body are
**not** detected by the markdown inline rule (requires the `{namespace}/{model-name}` slash form).
