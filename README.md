
# AIBoMGen CLI (WIP)

[![codecov](https://codecov.io/gh/idlab-discover/AIBoMGen-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/idlab-discover/AIBoMGen-cli)

Work-in-progress Go CLI that scans a repository for **basic Hugging Face model usage** and emits a **CycloneDX AI BOM (AIBOM)**.

## Status (WIP)

What works today:

- Basic scanning for Hugging Face model IDs in Python-like sources via `from_pretrained("...")`.
- AIBOM generation per detected model in JSON or XML.
- Hugging Face Hub API fetch to populate metadata fields.
- Hugging Face Repo README fetch to populate more metadata fields.
- Completeness scoring and validation of existing AIBOM files.
- Interactive or file based metadata enrichment.

What is future work:

- Improving the scanner beyond the current regex-based Hugging Face detection.
- Implementing data components with dataset fetchers and linking them in the aibom.
- Implementing the possibility to merge aiboms with existing sboms from a different source.
- Implementing the possibility to sign aiboms with cosign.

## Build

```bash
go test ./...
go build -o aibomgen-cli .
./aibomgen-cli --help
```

## Commands

### `generate`

Scans a directory for model usage and writes one AIBOM file per detected model.

```bash
./aibomgen-cli generate -i testdata/repo-basic
```

By default this writes JSON files under `dist/` with filenames derived from the model ID, e.g.:

- `dist/google-bert_bert-base-uncased_aibom.json`
- `dist/templates_model-card-example_aibom.json`

Common options:

- `--format json|xml|auto` (default: `auto`)
- `--output <path>`: the **directory portion** is used as output directory (default: `dist/aibom.json` → outputs to `dist/`)
- `--hf-mode online|dummy` (default: `online`)
- `--hf-token <token>` for gated/private models
- `--hf-timeout <seconds>`
- `--log-level quiet|standard|debug`

Experimental/stubbed:

- `--enrich`: attempts interactive completion, but the underlying enricher is not implemented yet.

### `validate`

Validates an existing AIBOM file (JSON/XML), runs completeness checks, and can fail in strict mode.

```bash
./aibomgen-cli validate -i dist/google-bert_bert-base-uncased_aibom.json
./aibomgen-cli validate -i dist/google-bert_bert-base-uncased_aibom.json --strict --min-score 0.5
```

Useful options:

- `--format json|xml|auto`
- `--strict` (fail on missing required fields)
- `--min-score 0.0-1.0`
- `--check-model-card` (default: `true`)
- `--log-level quiet|standard|debug`

### `completeness`

Computes and prints a completeness score for an existing AIBOM using the metadata field registry.

```bash
./aibomgen-cli completeness -i dist/google-bert_bert-base-uncased_aibom.json
```

Options:

- `--format json|xml|auto`
- `--log-level quiet|standard|debug`

### `enrich`

Command exists, but is currently not implemented.

```bash
./aibomgen-cli enrich --help
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json
```

### Global flags

- `--no-color`: disable ANSI coloring
- `--config <path>`: optional config file. If not provided, the app attempts to read a Viper config from the home directory (see `cmd/root.go`).

## Package overview

Each folder below is a Go package.

### `main`

Entry point that calls the Cobra root command.

### `cmd`

Cobra CLI wiring: root command, subcommands, flag parsing, and orchestration into `internal/*` packages.

### `internal/scanner`

Repository scanning.

- Current behavior: walks files and detects Hugging Face model IDs by regex matching `from_pretrained("<id>")` in `.py`, `.ipynb`, and `.txt`.
- Important limitation: weight-file detection is intentionally disabled right now.
- Future work: broaden detection beyond the current basic Hugging Face pattern.

### `internal/fetcher`

HTTP client for fetching model metadata from the Hugging Face Hub API (`/api/models/:id`).

- Used when `generate --hf-mode online`.
- Supports optional bearer token via `--hf-token`.

### `internal/metadata`

Central “field registry” describing which CycloneDX ML-BOM fields we care about.

- Defines keys, how to populate them (`Apply`), and how to check presence (`Present`).
- Used by `internal/builder` to populate the BOM and by `internal/completeness` to score it.

### `internal/builder`

Turns a scan result (and optional Hugging Face API response) into a CycloneDX BOM.

- Creates a minimal ML model component skeleton.
- Applies the `internal/metadata` registry once to populate fields.

### `internal/generator`

Orchestrates “per discovery” generation.

- For each detected model: fetch metadata (online mode) and build a BOM via the builder.
- Returns a list of generated BOMs back to the `generate` command.

### `internal/io`

Read/write helpers for CycloneDX BOMs.

- Supports JSON and XML.
- Supports `format=auto` based on file extension.
- Supports optional CycloneDX spec version selection for output.

### `internal/completeness`

Computes a completeness score $0..1$ for a BOM using weights defined in the metadata registry.

### `internal/validator`

Validates an existing AIBOM.

- Performs basic structural checks.
- Validates CycloneDX spec version.
- Runs completeness scoring and can enforce thresholds in strict mode.

### `internal/enricher`

Intended to interactively fill missing metadata fields.

- Current status: stubbed / not implemented.
- Future work: implement user prompting and (optionally) model card fetching.

### `internal/logging`

Small opt-in logger used across internal packages (writes only when a writer is configured).

### `internal/ui`

Very small ANSI-color helper used for banners and colored log prefixes.

## Docs and examples

- `testdata/repo-basic` is a small repository used in tests and examples.
- `docs/` contains design notes and mapping documentation.


