from securesystemslib.signer._key import SSlibKey
from securesystemslib.signer import CryptoSigner
from in_toto.models.layout import Layout, Step
from in_toto.models.metadata import Metablock
import os
import json


def generate_keys(output_dir):
    """
    Generate an Ed25519 key pair for the worker and save them to files.

    Args:
        output_dir (str): Directory to save the keys.
    Returns:
        CryptoSigner: A signer object for the worker's private key.
    """
    # Ensure the output directory exists
    os.makedirs(output_dir, exist_ok=True)

    # Generate Ed25519 key pair
    worker_signer = CryptoSigner.generate_ed25519()

    # Save the private key
    private_key_path = os.path.join(output_dir, "worker_private_key.pem")
    with open(private_key_path, "wb") as priv_file:
        priv_file.write(worker_signer.private_bytes)
    print(f"Private key saved to: {private_key_path}")

    # Save the public key as a JSON file
    public_key_path = os.path.join(output_dir, "worker_public_key.json")
    public_key_dict = worker_signer.public_key.to_dict()
    # Add the keyid field
    public_key_dict["keyid"] = worker_signer.public_key.keyid
    with open(public_key_path, "w") as pub_file:
        json.dump(public_key_dict, pub_file, indent=4)
    print(f"Public key saved to: {public_key_path}")

    return worker_signer


def load_signer(private_key_path, public_key_path):
    """
    Load the private key and public key to recreate a CryptoSigner object.

    Args:
        private_key_path (str): Path to the private key file.
        public_key_path (str): Path to the public key JSON file.

    Returns:
        CryptoSigner: A signer object for the private key.
    """
    # Load the public key from the JSON file
    with open(public_key_path, "r") as pub_file:
        public_key_dict = json.load(pub_file)
        public_key = SSlibKey.from_dict(
            public_key_dict["keyid"], public_key_dict)

    # Load the private key
    signer = CryptoSigner.from_priv_key_uri(
        f"file2://{private_key_path}", public_key)
    return signer


def create_layout(worker_signer, output_path):
    """
    Create an in-toto layout and save it to a JSON file.

    Args:
        worker_signer (CryptoSigner): The worker's signer object.
        output_path (str): Path to save the generated layout JSON file.
    Returns:
        Metablock: The created layout metadata.
    """
    # Create a new layout
    layout = Layout()
    layout.readme = "This layout defines the steps for the run_training task."
    # Set expiration to 90 days from now
    layout.set_relative_expiration(days=90)

    # Add the worker's public key to the layout
    worker_key_dict = worker_signer.public_key.to_dict()
    worker_key_dict["keyid"] = worker_signer.public_key.keyid
    layout.add_functionary_key(worker_key_dict)

    # Define the step for the `run_training` task
    step = Step(name="run_training")
    step.pubkeys = [worker_signer.public_key.keyid]
    step.set_expected_command_from_string("python tasks.py run_training")
    step.add_material_rule_from_string("MATCH model WITH PRODUCTS FROM api")
    step.add_material_rule_from_string("MATCH dataset WITH PRODUCTS FROM api")
    step.add_material_rule_from_string(
        "MATCH dataset_definition WITH PRODUCTS FROM api")
    step.add_product_rule_from_string("CREATE trained_model.keras")
    step.add_product_rule_from_string("CREATE metrics.json")

    # Add the step to the layout
    layout.steps = [step]

    # Wrap the layout in a Metablock
    layout_metadata = Metablock(signed=layout)

    # Save the layout to a JSON file
    layout_metadata.dump(output_path)
    print(f"Layout saved to: {output_path}")

    return layout_metadata


def sign_layout(layout_metadata, worker_signer, output_path):
    """
    Sign the in-toto layout with the worker's private key.

    Args:
        layout_metadata (Metablock): The layout metadata to sign.
        worker_signer (CryptoSigner): The worker's signer object.
        output_path (str): Path to save the signed layout JSON file.
    """
    # Sign the layout
    layout_metadata.create_signature(worker_signer)

    # Save the signed layout
    layout_metadata.dump(output_path)
    print(f"Signed layout saved to: {output_path}")


if __name__ == "__main__":
    # Directory to save keys
    secrets_dir = "./aibomgen-platform/secrets"
    os.makedirs(secrets_dir, exist_ok=True)

    # Generate worker keys
    signer = generate_keys(secrets_dir)

    # Define the path for the layout file
    layout_path = os.path.join(secrets_dir, "layout.json")

    # Create the layout and save it to the specified file
    layout_metadata = create_layout(signer, layout_path)

    # Sign the layout with the worker's private key
    signed_layout_path = os.path.join(secrets_dir, "signed_layout.json")
    sign_layout(layout_metadata, signer, signed_layout_path)
