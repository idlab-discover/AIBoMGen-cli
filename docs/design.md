## AIBoMGen Design (Initial Skeleton)

### Purpose
Generate an AI-aware Bill of Materials (AIBOM) that complements a traditional SBOM (e.g., Syft). Automatically detect AI-specific components (models, datasets, weight files, inference/inference-engine frameworks) and enrich them with metadata. Minimize user input; default operation scans the current project and outputs a CycloneDX-style JSON file.

### High-Level Workflow
1. Scan repository for AI artifacts (model IDs, weight files, AI frameworks).
2. Fetch metadata (Hugging Face model cards, local model card files) with caching/offline support.
3. Integrate an existing SBOM if present (Syft output) and enrich rather than duplicate.
4. Annotate components with vulnerability/trust signals (future phases).
5. Emit a merged CycloneDX JSON BOM (`dist/aibom.json` by default).

### Complement to Syft
Syft enumerates OS, language packages, container layers. AIBoMGen focuses on domain-specific assets Syft does not surface: model identities, weight provenance, dataset references, model card metadata, and trust indicators. When a Syft SBOM is present, AIBoMGen merges AI components and relationships into a combined BOM rather than generating duplicates.

### Automation Principles
- Zero required flags; sensible defaults.
- Auto-detect SBOM presence (`sbom.json`, `syft.json`, `bom.json`).
- Output directory defaults to `dist/` (created if missing).
- Offline mode uses local cache; network fetches avoided.

### Detection Heuristics (Initial Scope)
- Weight file extensions: `.pt`, `.pth`, `.bin`, `.safetensors`, `.ckpt`, `.onnx`, `.tflite`.
- Code patterns: `from_pretrained("<org>/<model>")`; huggingface hub calls.
- Framework presence in manifests: `transformers`, `torch`, `tensorflow`, `onnxruntime`, `diffusers`.
- Directory names: `models/`, `checkpoints/`, `weights/`.
- Confidence scoring: simple heuristic (e.g., path + pattern) to prepare for future filtering.

### Metadata Sources (Planned)
Primary: Hugging Face Hub API.
Secondary: Local `model-card.md` / `model-card.yaml` / `README.md` near weights.
Tertiary: Framework metadata (ONNX model metadata, PyTorch state dict info).
Caching: `~/.aibom/cache` (future).

### Merging Strategy (Planned)
- Ingest Syft CycloneDX or JSON output.
- Normalize components into CycloneDX representation.
- Enrich existing components when overlap (e.g., a weight file already listed).
- Add AI components with properties and external references; preserve provenance & detection evidence.

### CycloneDX Mapping (Initial Simplification)
Represent each AI artifact as a CycloneDX `component` with `type` set to `library` or a reserved value (future extension). Use `properties` for:
- `aibomgen.type` (model|weight-file|dataset|engine)
- `aibomgen.source` (path or URL)
- `aibomgen.confidence`
- `aibomgen.evidence` (pattern or file that triggered detection)
- Future: `aibomgen.modelCardURL`, `aibomgen.digest`, `aibomgen.license`, `aibomgen.trustSignals`

### CLI Commands (Planned)
- `generate`: scan, fetch metadata, merge SBOM (auto), write BOM.
  - Flags (initial subset implemented): `--path`, `--output`.
  - Future flags: `--offline`, `--include-sbom`, `--verbose`.
- `merge`: merge existing SBOM + AIBOM (future).
- `validate`: schema and completeness checks (future).

### Package Architecture (Current Skeleton)
- `internal/scanner`: detection logic (implemented minimally).
- `internal/fetcher`: metadata fetchers (placeholder).
- `internal/generator`: BOM assembly and writing (basic JSON writer now).
- Future: `internal/merger`, `internal/vuln`, `internal/models`.

### Data Model (Internal Representation)
`Component` (internal):
- `ID` (string)
- `Name` (string)
- `Type` (string) – model, weight-file, dataset, engine
- `Path` (string)
- `Evidence` (string)
- `Confidence` (float64)

Mapped directly into CycloneDX fields (simplified) with properties to retain custom attributes.

### Testing Strategy (Initial)
- Unit tests for scanner heuristics (regex + file extension detection).
- Future integration tests: merging with sample Syft output.
- Schema validation (CycloneDX) later once full generator implemented.

### Roadmap (Near-Term)
1. Expand scanner to parse manifests (requirements.txt, pyproject.toml).
2. Implement Hugging Face fetcher with caching & offline toggle.
3. Add merger for Syft SBOM enrichment.
4. Produce fully spec-compliant CycloneDX output.
5. Introduce trust & risk annotation layer.

### Non-Goals (Initial Phase)
- Full vulnerability scanning (delegated to existing tools for now).
- Proprietary model exfiltration (never upload weight contents).
- Complex risk scoring beyond basic placeholders.

### Security & Privacy Considerations
- Only outbound requests for public model metadata (unless token provided).
- Never transmit local file contents externally (just IDs / hashes future phase).
- Provide clear separation between detection evidence and user code contents.

### Acceptance Criteria (This Skeleton)
- `generate` command produces a JSON file with detected AI components (weight files + HF model IDs) into `dist/aibom.json`.
- Basic unit test validating detection of a HF model ID and weight file.
- Design document capturing architecture and roadmap (this file).
