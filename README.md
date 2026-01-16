# FamilyShare

Lightweight, self-hosted photo sharing for families on low-resource VPS.

This repository contains the backend and SSR frontend for FamilyShare. See `.docs/` for the Technical Design Document and task backlog.

Quick start:

1. Copy `.env.example` to `.env` and fill values.
2. Build and run:

```bash
go build -o familyshare ./cmd/app
./familyshare
```

For full setup instructions see the TDD in `.docs/familyshare-tdd.md`.
