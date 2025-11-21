# CRA oriented version (GO implementation)

## Summary

We need a consumer-ready (CRA-oriented) version of the AIBOM generator that targets systems embedding AI components (e.g., smart doorbells, IoT edge devices).
This version should assume trusted model developers and focus on lightweight integration, signed model metadata, and CycloneDX/SPDX compliant AIBOM output suitable for CI/CD, OTA updates or device inventory systems.

## Motivation

The current AIBOM generator targets high-assurance AI supply chain environments (e.g., research, enterprise, or regulated sectors).
For commercial or consumer devices, we need a version that:
- Works with trusted model sources (e.g., Hugging Face models).
- Generates minimal, verifiable AIBOMs without requiring full lineage tracking.
- Is easy to embed in firmware pipelines, cloud services, or edge deployments.
- Enables traceability of AI components while maintaining a low integration burden.

## Requirements

Functional:
- AIBOM schema focussed on AI components, runtime environment, and system dependencies.
- Suitable for OTA updates and embedded devices
- Assume model developer is trusted but require developer-provided signed metadata
- Automatic Hugging Face metadata retrieval
- CVE and vulnerability integration
  - Match dependencies with CVEs
  - Annotate AIBOM output with vulnerabilities and severity levels
  - Support integration with public vulnerability databases
- Easy integration for manufacturers
  - CLI and Python API for inclusion in build or CI/CD pipelines
  - Compatible with GitHub Actions, GitLab CI, and local builds
- Ofcourse CycloneDX and SPDX compliant output specification format

Non-functional:
- Minimal dependencies
- Runs offline (without metadata retrieval)
- JSON schema versioning and validation support
- Execution time as low as possible

## Example use case
A manufacturer builds a smart doorbell that performs object detection (e.g., detecting people, packages, animals).
They integrate a Hugging Face model (e.g., openai/fast-object-detector) and want to generate a CRA-style AIBOM automatically during CI/CD.
They want the build pipeline to automatically:

1. Fetch model metadata from Hugging Face.
2. Collect local system and firmware info.
3. Scan for vulnerabilities in dependencies.
4. Generate and archive an aibom.json.

Example repository structure:
```
pde-doorbell/
├── src/
│   ├── main.py
│   └── inference.py
├── manifests/
│   ├── system.json
│   └── runtime_requirements.txt
├── tools/
│   └── aibom-cra.py
├── .github/
│   └── workflows/
│       └── aibom.yml
└── dist/
    └── aibom.json (generated)
```
`manifests/system.json`
```
{
  "artifact_id": "pde-doorbell-v1.2.0",
  "producer": {
    "system_id": "doorbell-vendor-x",
    "component": "PDE-doorbell-firmware",
    "contact": "devops@vendorx.com"
  },
  "system_components": [
    ...
  ]
}

```

## Possible extensions
- Allow merging multiple Hugging Face models into one AIBOM
- Integrate model card trust scores? model card tools?
- Export combined SBOM + AIBOM bundle
- Enable offline caching of Hugging Face metadata

