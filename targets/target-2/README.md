# Test Fixture: target-2

This is a minimal repository to exercise the AIBoMGen scanner.

Contains:
- Python referencing Hugging Face models via `from_pretrained()`
- Weight files with typical extensions (`.bin`, `.onnx`)
- A simple `requirements.txt` listing common ML frameworks

Use it with:

```
# From repository root
# go run . generate --path targets/target-2 --output dist/aibom-test.json
```