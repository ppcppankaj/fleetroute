-- ============================================================
-- Migration 007: Alert Rules & Notification Infrastructure
-- ============================================================

-- Vehicle groups (prerequisite for rules)
CREATE TABLE IF NOT EXISTS vehicle_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    name        TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_vehicle_groups_tenant ON vehicle_groups(tenant_id) WHERE deleted_at IS NULL;

-- Vehicle group membership
CREATE TABLE IF NOT EXISTS vehicle_group_members (
    group_id   UUID NOT NULL REFERENCES vehicle_groups(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, vehicle_id)
);

-- Alert rules with JSONB condition trees
CREATE TABLE IF NOT EXISTS alert_rules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id),
    name             TEXT NOT NULL,
    description      TEXT,
    template_id      TEXT,                              -- built-in template used (if any)
    conditions       JSONB NOT NULL,                    -- {"op":"and","conditions":[{"field":"speed","op":"gt","value":80}]}
    severity         TEXT NOT NULL DEFAULT 'warning',   -- info | warning | critical
    cooldown_s       INT  NOT NULL DEFAULT 300,         -- min seconds between repeated triggers per device
    actions          JSONB NOT NULL DEFAULT '[]',       -- [{"type":"alert"},{"type":"webhook","url":"..."}]
    vehicle_group_id UUID REFERENCES vehicle_groups(id),-- NULL = applies to all tenant vehicles
    enabled          BOOL NOT NULL DEFAULT true,
    trigger_count    BIGINT NOT NULL DEFAULT 0,
    last_triggered   TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX idx_alert_rules_tenant_enabled ON alert_rules(tenant_id, enabled) WHERE deleted_at IS NULL;

-- Notification channels per tenant
CREATE TABLE IF NOT EXISTS notification_channels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id),
    name       TEXT NOT NULL,
    type       TEXT NOT NULL,   -- email | sms | push | webhook
    config     JSONB NOT NULL,  -- type-specific: {"address":"..."} or {"url":"...","secret":"..."}
    enabled    BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_notification_channels_tenant ON notification_channels(tenant_id) WHERE deleted_at IS NULL;

-- Notification delivery log
CREATE TABLE IF NOT EXISTS notification_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    alert_id     UUID REFERENCES alerts(id),
    channel_id   UUID REFERENCES notification_channels(id),
    channel_type TEXT NOT NULL,
    recipient    TEXT NOT NULL,   -- email address, phone number, FCM token, webhook URL
    subject      TEXT,
    body         TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',  -- pending | sent | failed
    attempt      INT  NOT NULL DEFAULT 1,
    error_msg    TEXT,
    sent_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notification_log_alert   ON notification_log(alert_id);
CREATE INDEX idx_notification_log_tenant  ON notification_log(tenant_id, created_at DESC);
CREATE INDEX idx_notification_log_status  ON notification_log(status) WHERE status = 'pending';

-- Rule trigger history for analytics
CREATE TABLE IF NOT EXISTS rule_trigger_history (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id    UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    device_id  UUID NOT NULL REFERENCES devices(id),
    tenant_id  UUID NOT NULL,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    event_data JSONB
);
CREATE INDEX idx_rule_trigger_rule   ON rule_trigger_history(rule_id, triggered_at DESC);
CREATE INDEX idx_rule_trigger_device ON rule_trigger_history(device_id, triggered_at DESC);

-- RLS
ALTER TABLE vehicle_groups          ENABLE ROW LEVEL SECURITY;
ALTER TABLE vehicle_group_members   ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules             ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channels   ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_log        ENABLE ROW LEVEL SECURITY;
ALTER TABLE rule_trigger_history    ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON vehicle_groups        USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON alert_rules           USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON notification_channels USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON notification_log      USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON rule_trigger_history  USING (tenant_id = current_setting('app.tenant_id')::UUID);
