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
-- alert_rules is initially created in 000006_trips_alerts.up.sql.
-- Add new columns needed for rules + notification workflow.
ALTER TABLE alert_rules
    ADD COLUMN IF NOT EXISTS description      TEXT,
    ADD COLUMN IF NOT EXISTS severity         TEXT NOT NULL DEFAULT 'warning',
    ADD COLUMN IF NOT EXISTS vehicle_group_id UUID,
    ADD COLUMN IF NOT EXISTS enabled          BOOL NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS trigger_count    BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_triggered   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at       TIMESTAMPTZ;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'alert_rules'
          AND column_name = 'is_active'
    ) THEN
        UPDATE alert_rules
        SET enabled = COALESCE(is_active, true)
        WHERE enabled IS DISTINCT FROM COALESCE(is_active, true);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE c.conname = 'alert_rules_vehicle_group_id_fkey'
          AND t.relname = 'alert_rules'
          AND n.nspname = 'public'
    ) THEN
        ALTER TABLE alert_rules
            ADD CONSTRAINT alert_rules_vehicle_group_id_fkey
            FOREIGN KEY (vehicle_group_id) REFERENCES vehicle_groups(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_enabled ON alert_rules(tenant_id, enabled) WHERE deleted_at IS NULL;

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
