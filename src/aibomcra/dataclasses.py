from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional


@dataclass
class GenerateOptions:
    model_id: str
    system_manifest: Optional[Path] = None
    requirements: Optional[Path] = None
    # support multiple CycloneDX inputs
    import_sbom: Optional[List[Path]] = None
    output: Optional[Path] = None
    fmt: str = "cyclonedx"  # or "spdx"
    offline: bool = False
    cache_dir: Optional[Path] = None
    prefer_imported: bool = False
