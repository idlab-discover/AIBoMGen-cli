from datasets import load_dataset


def load_datasets():
    # Load a sample dataset
    dataset = load_dataset("wmt14", "de-en")

    # Another dataset
    dataset2 = load_dataset("glue", "mrpc")

    return dataset, dataset2


if __name__ == "__main__":
    load_datasets()
