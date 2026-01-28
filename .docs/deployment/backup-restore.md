# Backup & Restore

## Backup
**SQLite database**
```
sqlite3 /opt/familyshare/data/familyshare.db ".backup /opt/familyshare/backups/familyshare.db"
```

**Photos**
```
tar -czf /opt/familyshare/backups/photos.tar.gz -C /opt/familyshare data/photos
```

**Environment**
```
cp /opt/familyshare/.env /opt/familyshare/backups/.env
```

## Restore
1. Stop containers:
   - `docker compose -f /opt/familyshare/deploy/docker-compose.yml down`
2. Restore database:
```
cp /opt/familyshare/backups/familyshare.db /opt/familyshare/data/familyshare.db
```
3. Restore photos:
```
tar -xzf /opt/familyshare/backups/photos.tar.gz -C /opt/familyshare
```
4. Restore `.env` if needed.
5. Start containers:
   - `docker compose -f /opt/familyshare/deploy/docker-compose.yml up -d`
