from __future__ import annotations

import json
from pathlib import Path

import pytest

from aibomcra.generator import (
    GenerateOptions,
    generate_aibom,
    load_system_manifest,
    parse_requirements,
    import_cyclonedx_sbom,
    validate_aibom,
    write_json,
)


def test_parse_requirements_pinned_and_unpinned(tmp_path: Path):
    req = tmp_path / "requirements.txt"
    req.write_text("""
    # comment
    numpy==1.26.4
    uvicorn
    
    requests == 2.31.0
    """.strip() + "\n", encoding="utf-8")

    deps = parse_requirements(req)
    assert {d["name"] for d in deps} >= {"numpy", "uvicorn", "requests"}
    assert any(d["name"] == "numpy" and d["version"] == "1.26.4" for d in deps)
    assert any(d["name"] == "uvicorn" and d["version"] == "" for d in deps)


def test_load_system_manifest_type_check(tmp_path: Path):
    # ok: dict
    ok = tmp_path / "sys.json"
    ok.write_text(json.dumps(
        {"system_components": [{"name": "ubuntu"}]}), encoding="utf-8")
    assert load_system_manifest(ok)["system_components"][0]["name"] == "ubuntu"

    # bad: list -> ValueError
    bad = tmp_path / "bad.json"
    bad.write_text(json.dumps([1, 2, 3]), encoding="utf-8")
    with pytest.raises(ValueError):
        load_system_manifest(bad)


def test_import_cyclonedx_sbom_merge_and_dedup(tmp_path: Path):
    sbom1 = tmp_path / "a.json"
    sbom2 = tmp_path / "b.json"
    sbom1.write_text(json.dumps({
        "components": [
            {"name": "numpy", "version": "1.26.4",
                "type": "library", "bom-ref": "pkg:numpy@1.26.4"},
            {"name": "torch", "version": "2.2.2", "type": "library"},
        ],
        "vulnerabilities": [
            {"id": "CVE-1", "affects": [{"ref": "pkg:numpy@1.26.4"}],
                "ratings": [{"severity": "HIGH", "score": 8.0}]}
        ]
    }), encoding="utf-8")
    sbom2.write_text(json.dumps({
        "components": [
            {"name": "numpy", "version": "1.26.4",
                "type": "library"},  # duplicate on purpose
            {"name": "urllib3", "version": "2.2.1", "type": "library"},
        ],
        "vulnerabilities": [
            {"id": "CVE-1", "affects": [{"ref": "pkg:numpy@1.26.4"}],
                "ratings": [{"severity": "HIGH", "score": 8.0}]}
        ]
    }), encoding="utf-8")

    comps, vulns = import_cyclonedx_sbom([sbom1, sbom2])
    names = [c["name"] for c in comps]
    assert names.count("numpy") == 1  # dedup component
    assert set(names) >= {"numpy", "torch", "urllib3"}

    # vulnerabilities deduped by (id, affects)
    ids = [v["id"] for v in vulns]
    assert ids.count("CVE-1") == 1


def test_generate_aibom_structure_minimal(tmp_path: Path):
    opts = GenerateOptions(
        model_id="org/model",
        system_manifest=None,
        requirements=None,
        import_sbom=None,
        output=tmp_path / "out.json",
    )
    aibom = generate_aibom(opts)
    ok, errors = validate_aibom(aibom)
    assert ok, f"unexpected errors: {errors}"
    assert aibom["metadata"]["model_id"] == "org/model"
    # model component present
    assert any(c["type"] == "ai-model" for c in aibom["components"])


def test_validate_aibom_errors():
    ok, errors = validate_aibom({"metadata": {}, "components": []})
    assert ok and errors == []

    ok2, errors2 = validate_aibom([1, 2, 3])  # type: ignore[arg-type]
    assert not ok2 and "AIBOM must be a JSON object" in errors2[0]


def test_write_json(tmp_path: Path):
    data = {"a": 1}
    out = tmp_path / "sub" / "out.json"
    write_json(data, out)
    assert out.exists()
    loaded = json.loads(out.read_text(encoding="utf-8"))
    assert loaded == data
