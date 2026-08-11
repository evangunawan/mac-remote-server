# Agent Guidelines for Mac Remote Server

## Technology Stack & Architecture
- **Language**: Go
- **Structure**: Clean Architecture / Domain-Driven Design (DDD).

## Versioning & Build Protocol
- Version is stored in `VERSION` file (e.g. `v1.0.1`) and git tags (`vX.Y.Z`).
- Binary is built using `./install.sh`, which populates build version into the binary via `-ldflags "-X main.Version=${VERSION}"`.
- When updating version:
  1. Update `VERSION` file.
  2. Run `./install.sh` to test compilation and local binary installation (`~/.local/bin/mac-remote-server`).
  3. Commit changes, create annotated git tag (`git tag -a vX.Y.Z -m "Release vX.Y.Z"`), and push with tags (`git push origin main --tags`).
