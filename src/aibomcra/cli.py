from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from aibomcra.dataclasses import GenerateOptions
from aibomcra.generator import (
    generate_aibom,
    validate_aibom,
    write_json,
)


def _add_common_generate_flags(p: argparse.ArgumentParser) -> None:
    # The common flags for "generate" command
    p.add_argument("--model-id", required=True,
                   help="Hugging Face model id (e.g., org/model)")
    p.add_argument("--system", type=Path, help="Path to system manifest JSON")
    p.add_argument("--requirements", type=Path,
                   help="Path to runtime requirements.txt")
    p.add_argument("--import-sbom", dest="import_sbom", type=Path, action="append",
                   help="Import multiple CycloneDX SBOM file(s) (e.g., from Trivy).")
    p.add_argument("--output", type=Path,
                   default=Path("dist/aibom.json"), help="Output AIBOM JSON path")
    p.add_argument("--format", dest="fmt",
                   choices=["cyclonedx", "spdx"], default="cyclonedx", help="Output format")
    p.add_argument("--offline", action="store_true",
                   help="Offline mode (use local manifests/cache only)")
    p.add_argument("--cache-dir", type=Path,
                   default=Path(".aibom-cache"), help="Cache directory")
    p.add_argument("--prefer-imported", action="store_true",
                   help="Prefer imported SBOM component data on conflicts")


def cmd_generate(args: argparse.Namespace) -> int:
    # Generate AIBOM command handler
    opts = GenerateOptions(
        model_id=args.model_id,
        system_manifest=args.system,
        requirements=args.requirements,
        import_sbom=args.import_sbom,
        output=args.output,
        fmt=args.fmt,
        offline=args.offline,
        cache_dir=args.cache_dir,
        prefer_imported=args.prefer_imported,
    )
    aibom = generate_aibom(opts)
    write_json(aibom, opts.output)
    print(f"AIBOM written to {opts.output}")
    return 0


def cmd_validate(args: argparse.Namespace) -> int:
    # Validate AIBOM command handler
    path: Path = args.input
    try:
        with path.open("r", encoding="utf-8") as f:
            data = json.load(f)
    except Exception as e:
        print(f"Failed to read {path}: {e}", file=sys.stderr)
        return 2
    ok, errors = validate_aibom(data)
    if ok:
        print(f"Validation OK: {path}")
        return 0
    print("Validation failed:")
    for e in errors:
        print(f"- {e}")
    return 1


def cmd_cache(args: argparse.Namespace) -> int:
    # Cache command handler (stub)
    d: Path = args.dir
    d.mkdir(parents=True, exist_ok=True)
    print(f"Cache ready: {d}")
    return 0


def cmd_verify(args: argparse.Namespace) -> int:
    # Verify command handler (stub)
    print("verify: not implemented in minimal scaffold (will support Ed25519/JWS)")
    return 0


def build_parser() -> argparse.ArgumentParser:
    # Build the top-level argument parser
    p = argparse.ArgumentParser(
        prog="aibom-cra", description="CRA-oriented AIBoM generator")
    sub = p.add_subparsers(dest="cmd", required=True)

    p_gen = sub.add_parser("generate", help="Generate an AIBOM JSON")
    _add_common_generate_flags(p_gen)
    p_gen.set_defaults(func=cmd_generate)

    p_val = sub.add_parser("validate", help="Validate an AIBOM JSON")
    p_val.add_argument("--input", type=Path, required=True,
                       help="Path to AIBOM JSON")
    p_val.set_defaults(func=cmd_validate)

    p_cache = sub.add_parser("cache", help="Prepare or manage cache")
    p_cache.add_argument("pull", nargs="?", help="(reserved)")
    p_cache.add_argument("--dir", type=Path, default=Path(".aibom-cache"))
    p_cache.set_defaults(func=cmd_cache)

    p_ver = sub.add_parser(
        "verify", help="Verify signed model metadata (stub)")
    p_ver.add_argument("--model-metadata", type=Path, required=False)
    p_ver.add_argument("--sig", type=Path, required=False)
    p_ver.add_argument("--pubkey", type=Path, required=False)
    p_ver.set_defaults(func=cmd_verify)

    return p


def main(argv: list[str] | None = None) -> int:
    # Main entry point
    # Build parser and dispatch commands
    parser = build_parser()
    args = parser.parse_args(argv)
    return args.func(args)


if __name__ == "__main__":  # pragma: no cover
    # Run the main function and exit
    raise SystemExit(main())
