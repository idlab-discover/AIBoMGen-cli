
# AIBoMGen CLI

[![Build + Unit Tests](https://github.com/idlab-discover/AIBoMGen-cli/actions/workflows/build.yml/badge.svg)](https://github.com/idlab-discover/AIBoMGen-cli/actions/workflows/build.yml) [![Scan Integration](https://github.com/idlab-discover/AIBoMGen-cli/actions/workflows/integration.yml/badge.svg)](https://github.com/idlab-discover/AIBoMGen-cli/actions/workflows/integration.yml)

Go CLI tool and packages that scan a repository for **basic Hugging Face model usage** and emit a **CycloneDX AI Bill of Materials (AIBOM)**.

## Status

What works today:

- `scan` command: scan a directory for AI imports and emit one AIBOM per detected model
- `generate` command: generate an AIBOM directly from one or more Hugging Face model IDs, or interactively browse models
- Hugging Face Hub API fetch to populate metadata fields
- Hugging Face Repo README fetch to populate more metadata fields
- Completeness scoring and validation of existing AIBOM files
- Interactive or file based metadata enrichment
- Data components with dataset fetchers and linking them in the AIBOM
- Updated the UI to utilise Charm libraries
- [BETA] Testing
- [BETA] Merge one or more AIBOMs with an existing SBOM from a different source (e.g., Syft, Trivy) into a single comprehensive BOM
- [BETA] Interactive Hugging Face model browsing
- [BETA] Complex regex-based Hugging Face detection


What is future work:

- Implementing check-vuln command to check AI vulnerability databases
- Implementing AIBOM generation based of model files not on Hugging Face

## Build

```bash
go test ./...
go build -o aibomgen-cli .
./aibomgen-cli --help
```

## Commands

### `scan`

Scans a directory for AI-related imports (e.g., Hugging Face model IDs) and writes one AIBOM file per detected model.

```bash
./aibomgen-cli scan -i targets/target-2
./aibomgen-cli scan -i targets/target-3 --format xml --hf-mode online
```

By default this writes JSON files under `dist/` with filenames derived from the model ID, e.g.:

- `dist/google-bert_bert-base-uncased_aibom.json`
- `dist/templates_model-card-example_aibom.json`

Options:

- `--input, -i <path>`: directory to scan (default: current directory; cannot be used with `--hf-mode=dummy`)
- `--output, -o <path>`: output file path (directory portion is used)
- `--format, -f json|xml|auto` (default: `auto`)
- `--spec <version>`: CycloneDX spec version for output (e.g., `1.4`, `1.5`, `1.6`)
- `--hf-mode online|dummy` (default: `online`)
- `--hf-token <token>`: for gated/private models
- `--hf-timeout <seconds>`
- `--log-level quiet|standard|debug`

### `generate`

Generates an AIBOM from one or more Hugging Face model IDs specified directly, or through an interactive model browser. Use `scan` instead when you want to detect models from a source directory.

```bash
./aibomgen-cli generate -m google-bert/bert-base-uncased
./aibomgen-cli generate -m gpt2 -m meta-llama/Llama-3.1-8B
./aibomgen-cli generate --interactive
```

Options:

- `--model-id, -m <id>`: Hugging Face model ID (can be specified multiple times or comma-separated)
- `--interactive`: open an interactive model selector (cannot be used with `--model-id`)
- `--output, -o <path>`: output file path (directory portion is used)
- `--format, -f json|xml|auto` (default: `auto`)
- `--spec <version>`: CycloneDX spec version for output (e.g., `1.4`, `1.5`, `1.6`)
- `--hf-mode online|dummy` (default: `online`)
- `--hf-token <token>`: for gated/private models
- `--hf-timeout <seconds>`
- `--log-level quiet|standard|debug`

### `validate`

Validates an existing AIBOM file (JSON/XML), runs completeness checks, and can fail in strict mode.

```bash
./aibomgen-cli validate -i dist/google-bert_bert-base-uncased_aibom.json
./aibomgen-cli validate -i dist/google-bert_bert-base-uncased_aibom.json --strict --min-score 0.5
```

Useful options:

- `--format, -f json|xml|auto`
- `--strict`: fail on missing required fields
- `--min-score 0.0-1.0`
- `--check-model-card`: validate model card fields (default: `false`)
- `--log-level quiet|standard|debug`

### `completeness`

Computes and prints a completeness score for an existing AIBOM using the metadata field registry.

```bash
./aibomgen-cli completeness -i dist/google-bert_bert-base-uncased_aibom.json
```

Options:

- `--format, -f json|xml|auto`
- `--plain-summary`: print a single-line machine-readable summary (no styling)
- `--log-level quiet|standard|debug`

### `enrich`

Enriches an existing AIBOM by filling missing metadata fields interactively or from a configuration file.

```bash
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json --strategy interactive
./aibomgen-cli enrich -i dist/google-bert_bert-base-uncased_aibom.json --strategy file --file config/enrichment.yaml
```

Options:

- `--input, -i <path>`: path to existing AIBOM (required)
- `--output, -o <path>`: output file path (default: overwrite input)
- `--format, -f json|xml|auto`: input BOM format
- `--output-format json|xml|auto`: output BOM format (default: same as input)
- `--spec <version>`: CycloneDX spec version for output
- `--strategy interactive|file` (default: `interactive`)
- `--file <path>`: enrichment config file for file-based enrichment (default: `./config/enrichment.yaml`)
- `--required-only`: only enrich required fields
- `--min-weight <float>`: minimum weight threshold for fields to enrich
- `--refetch`: refetch model metadata from Hugging Face Hub before enrichment
- `--no-preview`: skip preview before saving
- `--hf-token <token>`: Hugging Face API token (for refetch)
- `--hf-base-url <url>`: Hugging Face base URL (for refetch)
- `--hf-timeout <seconds>`: Hugging Face API timeout (for refetch)
- `--log-level quiet|standard|debug`

### `merge`

**[BETA]** Merges one or more AIBOMs with an existing SBOM from a different source (e.g., Syft, Trivy) into a single comprehensive BOM.

The SBOM's application metadata is preserved as the main component, while AI/ML model and dataset components from the AIBOM(s) are added to the components list.

```bash
# Example workflow: Combine software dependencies with AI components

# 1. Generate SBOM for software dependencies using Syft
syft scan . -o cyclonedx-json > sbom.json

# 2. Generate AIBOM for AI/ML components using AIBoMGen
./aibomgen-cli scan -i . -o aibom.json

# 3. Merge them into a comprehensive BOM
./aibomgen-cli merge --aibom aibom.json --sbom sbom.json -o merged.json

# 4. Merge multiple AIBOMs with one SBOM (for projects using multiple models)
./aibomgen-cli merge --aibom model1_aibom.json --aibom model2_aibom.json --sbom sbom.json -o merged.json
```

Options:

- `--aibom <path>`: Path to AIBOM file (can be specified multiple times, required)
- `--sbom <path>`: Path to SBOM file (required)
- `--output, -o <path>`: Output path for merged BOM (required)
- `--format, -f json|xml|auto`: Output format (default: `auto`)
- `--deduplicate`: Remove duplicate components based on BOM-ref (default: `true`)
- `--log-level quiet|standard|debug`

### Global flags

- `--config <path>`: config file to use (default: `$HOME/.aibomgen-cli.yaml` or `./config/defaults.yaml`)

The config file is a YAML file that sets default values for any command flag, so you don't have to repeat them on the command line. Keys are namespaced by command:

```yaml
scan:
  hf-token: "hf_..."
  hf-mode: "online"
  log-level: "debug"

validate:
  strict: true
  min-score: 0.5
```

Any flag not passed on the CLI falls back to the config file value. CLI flags always take precedence. See [`config/defaults.yaml`](config/defaults.yaml) for a full reference of all available keys.


## Package overview

Each folder below is a Go package.

### `main`

Entry point that calls the Cobra root command.

### `cmd`

Cobra CLI wiring: root command, subcommands, flag parsing, and orchestration into `internal/*` packages.

### `internal/scanner`

Repository scanning used by the `scan` command.

- Current behavior: walks files and detects Hugging Face model IDs by regex matching `from_pretrained("<id>")` in `.py`, `.ipynb`, and various config/script file types.
- Important limitation: weight-file detection is intentionally disabled right now.
- Future work: broaden detection beyond the current basic Hugging Face pattern.

### `internal/fetcher`

HTTP clients for fetching model and dataset metadata from the Hugging Face Hub.

- Fetches model metadata via API (`/api/models/:id`) and README (model cards)
- Fetches dataset metadata via API (`/api/datasets/:id`) and README (dataset cards)
- Used when `scan --hf-mode online`, `generate --hf-mode online`, or when enriching with `--refetch`
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
- Returns a list of generated BOMs back to the `scan` and `generate` commands.

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

- Provides rich, styled output for all commands (scan, generate, validate, completeness, enrich, merge)
- Implements workflow tracking with task progress indicators
- Defines a consistent color palette and text styles across the application
- Includes specialized UI components for each command:
  - `generate.go`: generation workflow with progress tracking
  - `validation.go`: validation results with colored status indicators
  - `completeness.go`: completeness scoring with visual field breakdown
  - `workflow.go`: task-based progress tracking
  - `progress.go`: spinner and progress indicators
  - `styles.go`: centralized styling and color definitions

### `internal/merger`

**[BETA]** BOM merging functionality for combining AIBOMs with SBOMs from other sources.

- Merges one or more AIBOMs with an SBOM while preserving the SBOM's application metadata
- The SBOM's metadata component (application) remains as the primary metadata component
- AI/ML model and dataset components from AIBOMs are added to the components list (not metadata)
- Supports component deduplication based on BOM-ref to avoid duplicates
- Intelligently merges dependencies, compositions, tools, and external references
- Handles dependency graph merging with proper conflict resolution
- Generates merge statistics (component counts, duplicates removed, AIBOMs merged)
- Used by the `merge` command to create comprehensive BOMs combining AI/ML and traditional software components

## Docs and examples

- `targets/target-2` is a small repository used in tests and examples.
- `docs/` contains design notes and mapping documentation.


