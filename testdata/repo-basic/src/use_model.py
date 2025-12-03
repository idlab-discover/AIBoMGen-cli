from transformers import AutoModel, AutoTokenizer


def load_models():
    # Single-segment model id
    m1 = AutoModel.from_pretrained("google-bert/bert-base-uncased")
    # Org/model id
    m2 = AutoModel.from_pretrained("facebook/opt-1.3b")
    m3 = AutoModel.from_pretrained("templates/model-card-example")
    return m1, m2, m3


if __name__ == "__main__":
    load_models()
