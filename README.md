
# AIBoMGen CRA (WIP)

[![codecov](https://codecov.io/gh/<OWNER>/<REPO>/branch/<BRANCH>/graph/badge.svg)](https://codecov.io/gh/<OWNER>/<REPO>)

Work-in-progress Go CLI to auto-detect AI artifacts (Hugging Face model IDs in Python and common weight files) and emit CycloneDX AIBOM. Designed for consumer/embedded pipelines with near-zero config. Can be merged with already generated SBOMs (for example with Syft).

## Current
- Command: `generate` (scans path, writes `dist/aibom.json`).
- Detects: `from_pretrained("<id>")` + weight file extensions.
- Test repo: `testdata/repo-basic`.

## Planned
- AI metadata fetch, full compliant CycloneDX BOM, SBOM merge, vulnerabilities.

## Usage
```bash
go build -o aibomgen-cra .
./aibomgen-cra generate --path testdata/repo-basic
```

See `docs/design.md` for roadmap details.


