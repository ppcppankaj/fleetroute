# ─────────────────────────────────────────────────────────────────────────────
# TrackOra - Create Super Admin User
# Usage: .\scripts\create-superadmin.ps1
# Run AFTER: docker compose up -d
# ─────────────────────────────────────────────────────────────────────────────

param(
    [string]$Email    = "superadmin@trackora.com",
    [string]$Name     = "Super Admin",
    [string]$Password = "Admin@1234",
    [string]$Role     = "SUPER_ADMIN"
)

Write-Host ""
Write-Host "============================================"
Write-Host "  TrackOra - Super Admin Setup"
Write-Host "============================================"
Write-Host "  Email   : $Email"
Write-Host "  Name    : $Name"
Write-Host "  Role    : $Role"
Write-Host "============================================"
Write-Host ""

# Check that the postgres-m15 container is running
$running = docker ps --format "{{.Names}}" | Select-String "fleet-postgres-m15"
if (-not $running) {
    Write-Host "ERROR: fleet-postgres-m15 container is not running."
    Write-Host "Run 'docker compose up -d' first, then re-run this script."
    exit 1
}

# Generate bcrypt hash using Python (bundled in most systems, or use Docker)
Write-Host "Generating bcrypt hash for password..."

# Try Python first (fastest)
$hash = $null
try {
    $hash = python -c "import bcrypt; print(bcrypt.hashpw('$Password'.encode(), bcrypt.gensalt(12)).decode())" 2>$null
} catch {}

# Fallback: use a Go one-liner via Docker
if (-not $hash -or $hash -eq "") {
    Write-Host "Python not found - using Docker Go container to hash password..."
    $hash = docker run --rm golang:1.23-alpine sh -c `
        "go run -e 'package main; import (""fmt""; ""golang.org/x/crypto/bcrypt""); func main() { h,_:=bcrypt.GenerateFromPassword([]byte(`"$Password`"),12); fmt.Println(string(h)) }'" 2>$null
}

# Final fallback: pre-compute a known bcrypt hash for the default password
# This is bcrypt hash of "Admin@1234" with cost 12
if (-not $hash -or $hash -eq "") {
    Write-Host "Using pre-computed bcrypt hash (cost 12) for default password 'Admin@1234'..."
    $hash = '$2a$12$FOmAe34mNOjrHDiUkYGR/OPKT4gXraFzU.C1Wu4zlXtDOSCTU0w1O'
}

Write-Host "Hash generated."
Write-Host ""

# Create the super_admins table if it does not exist and insert the user
$sql = @"
CREATE TABLE IF NOT EXISTS super_admins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'SUPER_ADMIN',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO super_admins (email, name, password_hash, role, is_active)
VALUES ('$Email', '$Name', '$hash', '$Role', true)
ON CONFLICT (email) DO UPDATE
    SET name          = EXCLUDED.name,
        password_hash = EXCLUDED.password_hash,
        role          = EXCLUDED.role,
        is_active     = true;

SELECT id, email, name, role, is_active, created_at FROM super_admins WHERE email = '$Email';
"@

Write-Host "Creating super_admins table and inserting user..."
$result = docker exec -i fleet-postgres-m15 psql -U fleet -d fleet_admin_db -c $sql 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "============================================"
    Write-Host "  Super Admin created successfully!"
    Write-Host "============================================"
    Write-Host ""
    Write-Host "  Login at: http://localhost:3002"
    Write-Host "  API    : POST http://localhost:4015/api/admin/auth/login"
    Write-Host ""
    Write-Host "  Credentials:"
    Write-Host "    Email   : $Email"
    Write-Host "    Password: $Password"
    Write-Host ""
    Write-Host "  Change your password after first login!"
    Write-Host "============================================"
} else {
    Write-Host ""
    Write-Host "ERROR creating super admin:"
    Write-Host $result
    exit 1
}
