from __future__ import annotations

import json
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any

from aibomcra.dataclasses import GenerateOptions


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def _read_json(p: Path) -> Any:
    with p.open("r", encoding="utf-8") as f:
        return json.load(f)


def load_system_manifest(p: Optional[Path]) -> Dict[str, Any]:
    if not p:
        return {}
    if not p.exists():
        raise FileNotFoundError(f"system manifest not found: {p}")
    data = _read_json(p)
    if not isinstance(data, dict):
        raise ValueError("system manifest must be a JSON object")
    return data


def parse_requirements(p: Optional[Path]) -> List[Dict[str, str]]:
    if not p:
        return []
    if not p.exists():
        raise FileNotFoundError(f"requirements not found: {p}")
    deps: List[Dict[str, str]] = []
    pat = re.compile(r"^\s*([A-Za-z0-9_.\-]+)\s*==\s*([A-Za-z0-9_.\-]+)\s*$")
    with p.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            m = pat.match(line)
            if m:
                deps.append(
                    {"name": m.group(1), "version": m.group(2), "type": "python"})
            else:
                # Best-effort parse: treat as name without version
                deps.append({"name": line, "version": "", "type": "python"})
    return deps


def import_cyclonedx_sbom(paths: Optional[List[Path]]) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:
    """Aggregate (components, vulnerabilities) from one or more CycloneDX JSON files.
    Components contain keys: name, version, bom_ref (if present), purl (if present), type.
    Vulnerabilities normalized minimally with id, affects (list of bom_ref), severity info if available.
    """
    if not paths:
        return [], []

    def _load_one(p: Path) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:
        if not p.exists():
            raise FileNotFoundError(f"CycloneDX SBOM not found: {p}")
        data = _read_json(p)
        comps_in: List[Dict[str, Any]] = data.get("components", []) or []
        vulns_in: List[Dict[str, Any]] = data.get("vulnerabilities", []) or []

        components: List[Dict[str, Any]] = []
        for c in comps_in:
            components.append({
                "name": c.get("name"),
                "version": c.get("version", ""),
                "bom_ref": c.get("bom-ref") or c.get("bom_ref"),
                "purl": c.get("purl"),
                "type": c.get("type", "library"),
                "source": "cyclonedx",
            })

        vulns: List[Dict[str, Any]] = []
        for v in vulns_in:
            vid = v.get("id") or v.get("bom-ref") or v.get("name")
            affects = []
            for aff in v.get("affects", []) or []:
                ref = aff.get("ref") or aff.get("bom-ref")
                if ref:
                    affects.append(ref)
            sev = None
            ratings = v.get("ratings") or []
            if ratings:
                # pick the highest severity rating
                sev = sorted(ratings, key=lambda r: r.get(
                    "score", 0), reverse=True)[0]
            vulns.append({
                "id": vid,
                "source": (v.get("source", {}) or {}).get("name", "trivy"),
                "severity": (sev or {}).get("severity"),
                "score": (sev or {}).get("score"),
                "affects": affects,
            })
        return components, vulns

    all_components: List[Dict[str, Any]] = []
    all_vulns: List[Dict[str, Any]] = []
    seen_comp_keys = set()

    for p in paths:
        comps, vulns = _load_one(p)
        for c in comps:
            key = (c.get("name"), c.get("version"), c.get("type"))
            if key not in seen_comp_keys:
                seen_comp_keys.add(key)
                all_components.append(c)
        all_vulns.extend(vulns)

    # Deduplicate vulnerabilities by (id, affects)
    dedup: Dict[Tuple[Any, Tuple[str, ...]], Dict[str, Any]] = {}
    for v in all_vulns:
        key = (v.get("id"), tuple(sorted(v.get("affects") or [])))
        if key not in dedup:
            dedup[key] = v
    all_vulns = list(dedup.values())

    return all_components, all_vulns


def _merge_components(primary: List[Dict[str, Any]], imported: List[Dict[str, Any]], prefer_imported: bool) -> List[Dict[str, Any]]:
    by_key: Dict[Tuple[str, str], Dict[str, Any]] = {}
    for c in primary:
        key = (c.get("name", ""), c.get("type", ""))
        by_key[key] = c
    for c in imported:
        key = (c.get("name", ""), c.get("type", ""))
        if key in by_key:
            if prefer_imported:
                by_key[key] = {**by_key[key], **{k: v for k,
                                                 v in c.items() if v not in (None, "")}}
        else:
            by_key[key] = c
    return sorted(by_key.values(), key=lambda x: (x.get("type", ""), x.get("name", "")))


def generate_aibom(options: GenerateOptions) -> Dict[str, Any]:
    system = load_system_manifest(options.system_manifest)
    reqs = parse_requirements(options.requirements)
    imported_components, imported_vulns = import_cyclonedx_sbom(
        options.import_sbom)

    # Model component (minimal placeholder)
    model_component = {
        "name": options.model_id or "<unknown-model>",
        "type": "ai-model",
        "source": "huggingface",  # logical default; not fetched here
    }

    runtime_components = [
        {"name": d["name"], "version": d.get(
            "version", ""), "type": "python", "source": "requirements"}
        for d in reqs
    ]

    system_components = []
    for comp in (system.get("system_components") or []):
        if isinstance(comp, dict):
            system_components.append({
                "name": comp.get("name"),
                "version": comp.get("version", ""),
                "type": comp.get("type", "os"),
                "source": "system-manifest",
            })

    # Merge with imported components (CycloneDX)
    merged_components = _merge_components(
        [model_component] + runtime_components + system_components,
        imported_components,
        options.prefer_imported,
    )

    aibom: Dict[str, Any] = {
        "metadata": {
            "tool": "aibom-cra",
            "version": "0.1.0",
            "generated": _now_iso(),
            "format": options.fmt,
            "artifact_id": system.get("artifact_id"),
            "producer": system.get("producer"),
            "model_id": options.model_id,
            "offline": options.offline,
        },
        "components": merged_components,
    }

    # Attach vulnerabilities (from imported SBOM only in this minimal scaffold)
    if imported_vulns:
        aibom["vulnerabilities"] = imported_vulns

    return aibom


def validate_aibom(aibom: Dict[str, Any]) -> Tuple[bool, List[str]]:
    errors: List[str] = []
    if not isinstance(aibom, dict):
        return False, ["AIBOM must be a JSON object"]
    md = aibom.get("metadata")
    comps = aibom.get("components")
    if not isinstance(md, dict):
        errors.append("metadata missing or not an object")
    if not isinstance(comps, list):
        errors.append("components missing or not a list")
    return len(errors) == 0, errors


def write_json(data: Dict[str, Any], out_path: Path) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with out_path.open("w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, sort_keys=True)
        f.write("\n")
