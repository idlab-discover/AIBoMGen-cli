from transformers import AutoModel, AutoTokenizer


def load_models():
    m1 = AutoModel.from_pretrained("mcpotato/42-eicar-street")
