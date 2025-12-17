from transformers import AutoModel, AutoTokenizer


def load_models():
    # Bert
    m1 = AutoModel.from_pretrained("google-bert/bert-base-uncased")
    # ModelCard example
    t1 = AutoModel.from_pretrained("templates/model-card-example")

    return m1, t1


if __name__ == "__main__":
    load_models()
