
# AIBoMGen CLI (WIP)

[![codecov](https://codecov.io/gh/idlab-discover/AIBoMGen-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/idlab-discover/AIBoMGen-cli)

Work-in-progress Go CLI that scans a repository for **basic Hugging Face model usage** and emits a **CycloneDX AI BOM (AIBOM)**.

## Status (WIP)

What works today:

- Basic scanning for Hugging Face model IDs in Python-like sources via `from_pretrained("...")`
- AIBOM generation per detected model in JSON or XML including correct dependencies and BOMrefs
- Hugging Face Hub API fetch to populate metadata fields
- Hugging Face Repo README fetch to populate more metadata fields
- Completeness scoring and validation of existing AIBOM files
- Interactive or file based metadata enrichment
- Data components with dataset fetchers and linking them in the AIBOM
- Updated the UI to utilise Charm libraries

What is future work:

- Improving the scanner beyond the current regex-based Hugging Face detection
- Implementing the possibility to merge AIBOMs with existing sboms from a different source
- Implementing the possibility to sign AIBOMs with cosign
- Implementing check-vuln command to check AI vulnerability databases
- Implementing AIBOM generation based of model files not on Hugging Face

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
- `--enrich`: enable interactive metadata enrichment after generation

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

Enriches an existing AIBOM by filling missing metadata fields interactively or from a configuration file.

```bash
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json --strategy interactive
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json --strategy file --config config/enrichment.yaml
```

Options:

- `--strategy interactive|file` (default: `interactive`)
- `--config <path>`: configuration file for file-based enrichment
- `--required-only`: only enrich required fields
- `--min-weight <float>`: minimum weight threshold for fields to enrich
- `--refetch`: refetch metadata from Hugging Face Hub
- `--no-preview`: skip preview before applying changes
- `--hf-token <token>`: Hugging Face API token
- `--log-level quiet|standard|debug`

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

HTTP clients for fetching model and dataset metadata from the Hugging Face Hub.

- Fetches model metadata via API (`/api/models/:id`) and README (model cards)
- Fetches dataset metadata via API (`/api/datasets/:id`) and README (dataset cards)
- Used when `generate --hf-mode online` or when enriching with `--refetch`
- Supports optional bearer token via `--hf-token` for gated/private resources
- Includes dummy implementations for offline/testing scenarios
- Provides markdown extraction utilities for parsing model and dataset cards

### `internal/metadata`

Central "field registry" describing which CycloneDX AI-BOM fields we care about.

- Defines field specifications for model components, dataset components, and Hugging Face properties
- Each field has a key, weight, required status, apply logic, and presence check
- Supports multiple field types: `ComponentKey`, `ModelCardKey`, `HFPropsKey`, and `DatasetKey`
- Used by `internal/builder` to populate the BOM and by `internal/completeness` to score it
- Used by `internal/enricher` to identify missing fields and apply new values
- Includes helpers for parsing and applying metadata from API responses and model/dataset cards

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

Interactively or automatically fills missing metadata fields in an existing AIBOM.

- Supports two strategies: `interactive` (prompts user for values) and `file` (reads from config)
- Can refetch metadata from Hugging Face Hub to fill known fields automatically
- Enriches both model components and dataset components
- Shows before/after preview with completeness scoring
- Integrates with the metadata field registry to identify and fill missing fields
- Respects field weights and required status when prompting

### `internal/ui`

Comprehensive TUI (Terminal User Interface) system built with Charm libraries (Lipgloss, Bubbletea concepts).

- Provides rich, styled output for all commands (generate, validate, completeness, enrich)
- Implements workflow tracking with task progress indicators
- Defines a consistent color palette and text styles across the application
- Includes specialized UI components for each command:
  - `generate.go`: generation workflow with progress tracking
  - `validation.go`: validation results with colored status indicators
  - `completeness.go`: completeness scoring with visual field breakdown
  - `workflow.go`: task-based progress tracking
  - `progress.go`: spinner and progress indicators
  - `styles.go`: centralized styling and color definitions

## Docs and examples

- `testdata/repo-basic` is a small repository used in tests and examples.
- `docs/` contains design notes and mapping documentation.


