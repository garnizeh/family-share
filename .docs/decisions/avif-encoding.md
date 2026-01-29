# Decision: AVIF Encoding Support

## Date
2026-01-29

## Context
The database schema allows `photos.format` values of `webp` and `avif`, but the pipeline only produced WebP. We need AVIF encoding support without adding CGO dependencies, and keep the MVP stable for low-resource environments.

## Decision
Adopt `github.com/gen2brain/avif` for AVIF encode/decode. It is CGO-free, preferring native shared libraries when available and falling back to WASM via wazero. Add an `IMAGE_FORMAT` configuration option to select output format (`webp` default, `avif` optional).

## Consequences
- Default behavior remains WebP for compatibility and predictable performance.
- AVIF can be enabled by setting `IMAGE_FORMAT=avif`.
- Upload validation and pipeline can decode and encode AVIF.
- Schema and pipeline formats are aligned (`webp` and `avif`).
