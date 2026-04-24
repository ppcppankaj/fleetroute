CREATE TABLE alert_rules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    name       VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- e.g. GEOFENCE_BREACH, SPEEDING, MAINTENANCE_DUE, OFFLINE
    severity   VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    conditions JSONB NOT NULL,
    is_active  BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE active_alerts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    rule_id      UUID REFERENCES alert_rules(id),
    vehicle_id   UUID NOT NULL,
    driver_id    UUID,
    type         VARCHAR(50) NOT NULL,
    severity     VARCHAR(20) NOT NULL,
    message      TEXT NOT NULL,
    metadata     JSONB,
    status       VARCHAR(20) DEFAULT 'TRIGGERED', -- TRIGGERED, ACKNOWLEDGED, RESOLVED
    resolved_by  UUID,
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_alerts_tenant_status ON active_alerts(tenant_id, status);
