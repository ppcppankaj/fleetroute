CREATE TABLE super_admins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    name          VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50) DEFAULT 'ADMIN',  -- ADMIN | SUPER_ADMIN | SUPPORT
    is_active     BOOLEAN DEFAULT true,
    last_login    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE support_tickets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    subject     VARCHAR(255) NOT NULL,
    body        TEXT NOT NULL,
    status      VARCHAR(30) DEFAULT 'OPEN',  -- OPEN, IN_PROGRESS, RESOLVED, CLOSED
    priority    VARCHAR(20) DEFAULT 'MEDIUM',
    assigned_to UUID REFERENCES super_admins(id),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE platform_announcements (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      VARCHAR(255) NOT NULL,
    body       TEXT NOT NULL,
    target     VARCHAR(30) DEFAULT 'ALL',  -- ALL | TENANT | PLAN
    published  BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
