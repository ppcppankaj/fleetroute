CREATE TABLE tenants (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    slug         VARCHAR(100) UNIQUE NOT NULL,
    plan_id      VARCHAR(50),
    status       VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, TRIAL
    branding     JSONB,
    feature_flags JSONB,
    max_vehicles INT DEFAULT 10,
    max_users    INT DEFAULT 5,
    timezone     VARCHAR(50) DEFAULT 'UTC',
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE subscription_plans (
    id          VARCHAR(50) PRIMARY KEY,  -- e.g. "starter", "pro", "enterprise"
    name        VARCHAR(100) NOT NULL,
    price       NUMERIC(10,2) NOT NULL,
    currency    VARCHAR(3) DEFAULT 'USD',
    max_vehicles INT,
    max_users   INT,
    features    JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
