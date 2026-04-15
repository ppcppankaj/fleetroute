-- ============================================================
-- Migration 009: Async Report Job Queue & API Keys
-- ============================================================

-- Report job queue
CREATE TABLE IF NOT EXISTS report_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    requested_by    UUID NOT NULL REFERENCES users(id),
    report_type     TEXT NOT NULL,   -- trip | idle | fuel | driver_behavior | geofence_violations | overspeed | maintenance | ais140_audit
    title           TEXT,
    parameters      JSONB NOT NULL DEFAULT '{}',   -- {vehicle_ids:[], group_id:null, from:"2024-01-01", to:"2024-01-31", format:"csv"}
    format          TEXT NOT NULL DEFAULT 'csv',   -- csv | pdf
    status          TEXT NOT NULL DEFAULT 'pending', -- pending | processing | completed | failed
    progress_pct    SMALLINT NOT NULL DEFAULT 0,
    output_url      TEXT,       -- S3 pre-signed or permanent URL to download
    output_size_b   BIGINT,
    error_msg       TEXT,
    nats_message_id TEXT,       -- JetStream message ID for deduplication
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,  -- when output URL expires
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_report_jobs_tenant  ON report_jobs(tenant_id, created_at DESC);
CREATE INDEX idx_report_jobs_status  ON report_jobs(status) WHERE status IN ('pending','processing');
CREATE INDEX idx_report_jobs_user    ON report_jobs(requested_by, created_at DESC);

-- Scheduled reports (daily/weekly/monthly delivery)
CREATE TABLE IF NOT EXISTS scheduled_reports (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    created_by   UUID NOT NULL REFERENCES users(id),
    report_type  TEXT NOT NULL,
    title        TEXT NOT NULL,
    parameters   JSONB NOT NULL DEFAULT '{}',
    format       TEXT NOT NULL DEFAULT 'csv',
    schedule     TEXT NOT NULL,  -- daily | weekly | monthly
    recipients   JSONB NOT NULL DEFAULT '[]',  -- ["email@example.com"]
    enabled      BOOL NOT NULL DEFAULT true,
    last_run_at  TIMESTAMPTZ,
    next_run_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_scheduled_reports_tenant  ON scheduled_reports(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_scheduled_reports_next    ON scheduled_reports(next_run_at) WHERE enabled AND deleted_at IS NULL;

-- API keys for integration access
CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    created_by   UUID NOT NULL REFERENCES users(id),
    name         TEXT NOT NULL,
    key_hash     TEXT NOT NULL UNIQUE,  -- bcrypt hash of the actual key
    key_prefix   TEXT NOT NULL,         -- first 8 chars for display (e.g. "gps_abc1")
    permissions  JSONB NOT NULL DEFAULT '["read"]',   -- ["read","write","commands"]
    rate_limit   INT NOT NULL DEFAULT 1000,   -- requests per minute
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    enabled      BOOL NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_api_keys_tenant     ON api_keys(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix) WHERE deleted_at IS NULL;

-- Webhook endpoints
CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    name         TEXT NOT NULL,
    url          TEXT NOT NULL,
    secret       TEXT,          -- HMAC-SHA256 signing secret
    event_types  JSONB NOT NULL DEFAULT '["alert","trip","geofence"]',
    enabled      BOOL NOT NULL DEFAULT true,
    failure_count INT NOT NULL DEFAULT 0,
    last_delivery_at TIMESTAMPTZ,
    last_delivery_status TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX idx_webhook_endpoints_tenant ON webhook_endpoints(tenant_id) WHERE deleted_at IS NULL;

-- Webhook delivery log
CREATE TABLE IF NOT EXISTS webhook_delivery_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id    UUID NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    event_type     TEXT NOT NULL,
    payload        JSONB NOT NULL,
    http_status    INT,
    response_body  TEXT,
    attempt        INT NOT NULL DEFAULT 1,
    duration_ms    INT,
    delivered_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhook_delivery_endpoint ON webhook_delivery_log(endpoint_id, created_at DESC);

-- Raw packet diagnostic log (for GoAdmin inspector)
CREATE TABLE IF NOT EXISTS packet_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id      UUID REFERENCES devices(id),
    imei           TEXT,
    source_ip      TEXT NOT NULL,
    source_port    INT NOT NULL,
    protocol       TEXT NOT NULL,
    packet_size_b  INT NOT NULL,
    crc_ok         BOOL NOT NULL DEFAULT true,
    parse_ok       BOOL NOT NULL DEFAULT true,
    record_count   INT NOT NULL DEFAULT 0,
    parse_error    TEXT,
    raw_hex        TEXT,         -- truncated to 4096 chars for storage
    received_at    TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (received_at);

-- Create initial partition (current month)
CREATE TABLE packet_log_2026_04 PARTITION OF packet_log
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE packet_log_2026_05 PARTITION OF packet_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE packet_log_2026_06 PARTITION OF packet_log
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE INDEX idx_packet_log_imei  ON packet_log(imei, received_at DESC);
CREATE INDEX idx_packet_log_error ON packet_log(parse_ok, received_at DESC) WHERE NOT parse_ok;

-- RLS
ALTER TABLE report_jobs         ENABLE ROW LEVEL SECURITY;
ALTER TABLE scheduled_reports   ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys            ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_endpoints   ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON report_jobs       USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON scheduled_reports USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON api_keys          USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON webhook_endpoints USING (tenant_id = current_setting('app.tenant_id')::UUID);
