-- 000011_billing_in_main_db.up.sql
-- Replicate billing tables into the main DB so api-service can query them directly.
-- In production, prefer a dedicated billing service with its own DB.

CREATE TABLE IF NOT EXISTS subscription_plans (
    id          VARCHAR(50) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    price       NUMERIC(10,2) NOT NULL,
    currency    VARCHAR(3) DEFAULT 'INR',
    max_vehicles INT,
    max_users   INT,
    features    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO subscription_plans (id, name, price, currency, max_vehicles, max_users, features)
VALUES
  ('starter',    'Starter',       999,  'INR', 25,  5,  '{"video":false,"adas":false}'),
  ('pro',        'Professional',  2999, 'INR', 100, 20, '{"video":true,"adas":false}'),
  ('enterprise', 'Enterprise',    0,    'INR', -1,  -1, '{"video":true,"adas":true}')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID UNIQUE NOT NULL REFERENCES tenants(id),
    plan_id              VARCHAR(50) NOT NULL REFERENCES subscription_plans(id),
    stripe_sub_id        VARCHAR(100),
    stripe_cus_id        VARCHAR(100),
    status               VARCHAR(30) DEFAULT 'ACTIVE',
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id),
    stripe_inv_id VARCHAR(100),
    amount        NUMERIC(10,2) NOT NULL,
    currency      VARCHAR(3) DEFAULT 'INR',
    status        VARCHAR(20) DEFAULT 'OPEN',
    due_date      DATE,
    paid_at       TIMESTAMPTZ,
    pdf_url       VARCHAR(500),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices(tenant_id, created_at DESC);

ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices       ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON subscriptions USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);
CREATE POLICY tenant_isolation ON invoices       USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID);
