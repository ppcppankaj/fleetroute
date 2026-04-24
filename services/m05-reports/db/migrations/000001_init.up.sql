CREATE TABLE report_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(150) NOT NULL,
    type        VARCHAR(50) NOT NULL, -- TRIP_SUMMARY, FUEL, DRIVER_SCORE, MAINTENANCE, ALERT
    parameters  JSONB,
    schedule    VARCHAR(50), -- DAILY, WEEKLY, MONTHLY or null for on-demand
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE report_runs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    definition_id UUID REFERENCES report_definitions(id) ON DELETE CASCADE,
    tenant_id    UUID NOT NULL,
    status       VARCHAR(20) DEFAULT 'PENDING', -- PENDING, RUNNING, DONE, FAILED
    file_url     VARCHAR(500),
    started_at   TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error        TEXT
);
CREATE INDEX idx_runs_tenant ON report_runs(tenant_id, started_at DESC);
