
# AIBoMGen CLI (WIP)

Work-in-progress Go CLI to auto-detect AI artifacts (Hugging Face model IDs in Python and common weight files) and emit CycloneDX AIBOM. Designed for consumer/embedded pipelines with near-zero config. Can be merged with already generated SBOMs (for example with Syft).

## Current
- Command: `generate` (scans path, writes `dist/aibom.json`).
- Detects: `from_pretrained("<id>")` + weight file extensions.
- Test repo: `testdata/repo-basic`.

## Planned
- AI metadata fetch, full compliant CycloneDX BOM, SBOM merge, vulnerabilities.

## Usage
```bash
go build ./cmd/aibomgen-cli
./aibomgen-cli generate --path testdata/repo-basic
```

See `docs/design.md` for roadmap details.


