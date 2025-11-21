from __future__ import annotations

import json
from pathlib import Path

from aibomcra.cli import main
from aibomcra.generator import GenerateOptions, generate_aibom, write_json


def test_cli_generate_writes_output(tmp_path: Path):
    out = tmp_path / "aibom.json"
    rc = main([
        "generate",
        "--model-id", "org/model",
        "--output", str(out),
    ])
    assert rc == 0
    assert out.exists()
    data = json.loads(out.read_text(encoding="utf-8"))
    assert data["metadata"]["model_id"] == "org/model"


def test_cli_validate_ok_and_fail(tmp_path: Path):
    # ok file
    out = tmp_path / "ok.json"
    aibom = generate_aibom(GenerateOptions(model_id="org/model"))
    write_json(aibom, out)
    rc_ok = main(["validate", "--input", str(out)])
    assert rc_ok == 0

    # invalid JSON -> rc 2
    bad = tmp_path / "bad.json"
    bad.write_text("{not: json}", encoding="utf-8")
    rc_bad = main(["validate", "--input", str(bad)])
    assert rc_bad == 2

    # structurally invalid -> rc 1
    invalid = tmp_path / "invalid.json"
    invalid.write_text(json.dumps([1, 2, 3]), encoding="utf-8")
    rc_inv = main(["validate", "--input", str(invalid)])
    assert rc_inv == 1


def test_cli_cache_and_verify(tmp_path: Path):
    d = tmp_path / ".cache"
    rc_cache = main(["cache", "--dir", str(d)])
    assert rc_cache == 0 and d.exists()

    rc_verify = main(["verify"])  # stub
    assert rc_verify == 0
