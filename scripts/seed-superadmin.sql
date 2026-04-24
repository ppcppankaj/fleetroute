-- ─────────────────────────────────────────────────────────────────────────────
-- TrackOra - Super Admin seed
-- Password: Admin@1234  (bcrypt cost 12)
-- Change the password hash after first login via the API
-- ─────────────────────────────────────────────────────────────────────────────

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

-- bcrypt hash of 'Admin@1234' (cost 12)
INSERT INTO super_admins (email, name, password_hash, role, is_active)
VALUES (
    'superadmin@trackora.com',
    'Super Admin',
    '$2a$12$FOmAe34mNOjrHDiUkYGR/OPKT4gXraFzU.C1Wu4zlXtDOSCTU0w1O',
    'SUPER_ADMIN',
    true
)
ON CONFLICT (email) DO UPDATE
    SET name          = EXCLUDED.name,
        password_hash = EXCLUDED.password_hash,
        is_active     = true;

SELECT id, email, name, role, is_active, created_at FROM super_admins;
