#!/bin/sh
set -e

# Ensure dirs exist and have permissive group write so uploads work when host mounts
: "${DATA_DIR:=/app/data}"
: "${TEMP_UPLOAD_DIR:=/app/tmp_uploads}"

mkdir -p "$DATA_DIR" "$TEMP_UPLOAD_DIR"

# setgid so new files inherit group; allow group write
chmod 2775 "$DATA_DIR" "$TEMP_UPLOAD_DIR" || true

# If explicit UID/GID are provided, chown the dirs so the host user can also access them
if [ -n "$APP_UID" ] && [ -n "$APP_GID" ]; then
  chown -R "$APP_UID:$APP_GID" "$DATA_DIR" "$TEMP_UPLOAD_DIR" || true
fi

exec /app/familyshare
