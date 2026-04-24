# TrackOra - Windows secrets bootstrap
# Run ONCE before first: docker compose up -d
# Requires: OpenSSL (comes with Git for Windows)

$SecretsDir = ".\secrets"
New-Item -ItemType Directory -Force $SecretsDir | Out-Null

if (Test-Path "$SecretsDir\jwt_private.pem") {
    Write-Host "Secrets already exist - skipping generation."
    exit 0
}

Write-Host "Generating RSA-2048 JWT key pair..."
openssl genrsa -out "$SecretsDir\jwt_private.pem" 2048
if ($LASTEXITCODE -ne 0) { Write-Host "ERROR: openssl failed. Install Git for Windows (includes openssl)."; exit 1 }

openssl rsa -in "$SecretsDir\jwt_private.pem" -pubout -out "$SecretsDir\jwt_public.pem"

Write-Host ""
Write-Host "Keys generated successfully:"
Write-Host "   Private: $SecretsDir\jwt_private.pem"
Write-Host "   Public : $SecretsDir\jwt_public.pem"
Write-Host ""
Write-Host "WARNING: Never commit these files to git."
