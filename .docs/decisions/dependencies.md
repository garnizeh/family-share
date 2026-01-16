# Dependency Decisions — FamilyShare MVP

This document records the rationale for core dependencies chosen for the FamilyShare MVP. The guiding principles are: no CGO, minimal and well-maintained libraries, storage/CPU efficiency, and small dependency surface.

## Selected dependencies

- modernc.org/sqlite
  - Reason: pure-Go, no CGO required. Keeps deployment simple on low-resource VPS and avoids linking C libs.
  - Use: SQLite driver for application metadata storage.

- github.com/go-chi/chi/v5
  - Reason: lightweight, idiomatic router with good middleware support. Minimal overhead compared to heavier frameworks.
  - Use: HTTP routing and middleware chaining.

- github.com/disintegration/imaging
  - Reason: simple API for image resizing and basic processing; good performance and widely used.
  - Use: resizing and resampling images in the pipeline.

- github.com/chai2010/webp
  - Reason: pure-Go WebP encoder/decoder; stable option for producing efficient WebP files.
  - Use: encode processed images to WebP for storage and serving.

- github.com/rwcarlsen/goexif/exif
  - Reason: straightforward EXIF parsing for orientation and metadata extraction.
  - Use: read EXIF orientation tag and correct image orientation before resizing.

- golang.org/x/crypto/bcrypt
  - Reason: battle-tested password hashing library (Go's x/crypto). Sufficient for single-admin password use.
  - Use: verify admin password against stored bcrypt hash.

- github.com/joho/godotenv
  - Reason: convenience in local development to load `.env` files. Optional in production.
  - Use: load environment variables during development/test runs.


## Notes on optional/deferrable libraries
- AVIF support: a pure-Go AVIF encoder/decoder exists in some libraries, but AVIF tooling in Go is less mature and can be heavier. For the MVP we will produce WebP images and store an AVIF variant only if a stable pure-Go encoder is available and the benefit justifies added complexity.

- Session stores and advanced middleware (e.g., gorilla/sessions, Redis) are deferred for MVP. We will implement a simple SQLite-backed session table for persistence and keep middleware light. If later a distributed session store is required, we can introduce it.

## Security & Licensing
- All libraries chosen are open-source; verify licenses prior to release. Document any license concerns here.

## How to add/update dependencies
- Use `go get` to add a package, then `go mod tidy` to clean the module file and fetch transitive dependencies.
- Keep the dependency list minimal; add features only when necessary for MVP.
