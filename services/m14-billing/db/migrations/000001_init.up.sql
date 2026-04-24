CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID UNIQUE NOT NULL,
    plan_id         VARCHAR(50) NOT NULL,
    stripe_sub_id   VARCHAR(100),
    stripe_cus_id   VARCHAR(100),
    status          VARCHAR(30) DEFAULT 'ACTIVE', -- ACTIVE, PAST_DUE, CANCELLED
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE invoices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    stripe_inv_id  VARCHAR(100),
    amount         NUMERIC(10,2) NOT NULL,
    currency       VARCHAR(3) DEFAULT 'USD',
    status         VARCHAR(20) DEFAULT 'OPEN', -- OPEN, PAID, VOID, UNCOLLECTIBLE
    due_date       DATE,
    paid_at        TIMESTAMPTZ,
    pdf_url        VARCHAR(500),
    created_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_invoices_tenant ON invoices(tenant_id, created_at DESC);
