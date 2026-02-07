# Docker Compose Deployment (Caddy + FamilyShare)

## Prerequisites
- Docker Engine + Docker Compose v2 installed on the VPS
- A domain name pointing to the VPS
- Ports 80 and 443 open

## Directory Layout
```
/opt/familyshare
  ├─ .env
  ├─ data/
  ├─ tmp_uploads/
  ├─ deploy/
  │   ├─ docker-compose.yml
  │   └─ Caddyfile
```

## Steps
1. Create directories:
   - `/opt/familyshare/data`
   - `/opt/familyshare/tmp_uploads`
   - `/opt/familyshare/deploy`
2. Copy `deploy/docker-compose.yml` and `deploy/Caddyfile` into `/opt/familyshare/deploy` (optional — the repo already contains them).
3. Create `/opt/familyshare/.env` (see `.env.example`).
4. Set the required Caddy variables in `/opt/familyshare/.env`:
   - `DOMAIN=familyshare.example.com`
   - `ACME_EMAIL=admin@example.com`
5. Start using the project deploy script (recommended):

```bash
# From the project root
./scripts/deploy.sh
```

If you prefer to run Docker Compose manually, make sure the `.env` values are set and then run `docker compose -f /opt/familyshare/deploy/docker-compose.yml up -d`.

## Notes
- The app listens on `:8080` inside the Docker network.
- Data and SQLite DB live under `/opt/familyshare/data`.
- Temp uploads are stored under `/opt/familyshare/tmp_uploads`.
- Log rotation is configured via Docker logging options in `docker-compose.yml`.
