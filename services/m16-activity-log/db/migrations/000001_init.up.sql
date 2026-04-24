CREATE TABLE activity_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID,
    vehicle_id  UUID,
    driver_id   UUID,
    type        VARCHAR(100) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_activity_tenant ON activity_events(tenant_id, created_at DESC);
CREATE INDEX idx_activity_vehicle ON activity_events(vehicle_id, created_at DESC) WHERE vehicle_id IS NOT NULL;
