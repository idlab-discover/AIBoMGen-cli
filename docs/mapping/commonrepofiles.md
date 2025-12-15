# Common files in the repository

We can query model files such as safetensors, hash them and include them in BOM.

GET https://huggingface.co/google-bert/bert-base-uncased/resolve/main/:file

file can be: 
- LICENSE
- README.md (see previous)
- config.json (see previous)
- model.onnx
- model.safetensors...
- pytorch_model.bin
- rust_model.ot
- tf_model.h5
- tokenizer.json
- tokenizer_config.json
- vocab.txt