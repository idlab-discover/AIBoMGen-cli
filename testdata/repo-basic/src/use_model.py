from transformers import AutoModel, AutoTokenizer


def load_models():
    # Single-segment model id
    m1 = AutoModel.from_pretrained("bert-base-uncased")
    # Org/model id
    m2 = AutoModel.from_pretrained("facebook/opt-1.3b")
    t = AutoTokenizer.from_pretrained("bert-base-uncased")
    return m1, m2, t


if __name__ == "__main__":
    load_models()
