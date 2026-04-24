#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# TrackOra — One-time secrets bootstrap
# Run ONCE before first `docker compose up -d`
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SECRETS_DIR="./secrets"
mkdir -p "$SECRETS_DIR"

if [[ -f "$SECRETS_DIR/jwt_private.pem" ]]; then
  echo "✅  Secrets already exist — skipping generation."
  exit 0
fi

echo "🔑  Generating RSA-2048 JWT key pair..."
openssl genrsa -out "$SECRETS_DIR/jwt_private.pem" 2048
openssl rsa -in "$SECRETS_DIR/jwt_private.pem" -pubout -out "$SECRETS_DIR/jwt_public.pem"

chmod 600 "$SECRETS_DIR/jwt_private.pem"
chmod 644 "$SECRETS_DIR/jwt_public.pem"

echo ""
echo "✅  Keys generated:"
echo "   Private: $SECRETS_DIR/jwt_private.pem"
echo "   Public : $SECRETS_DIR/jwt_public.pem"
echo ""
echo "⚠️   NEVER commit these files to git."
echo "    They are in .gitignore by default."
