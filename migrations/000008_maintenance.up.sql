-- ============================================================
-- Migration 008: Maintenance & Vehicle Document Management
-- ============================================================

-- Service schedules: time-based and odometer-based
CREATE TABLE IF NOT EXISTS service_schedules (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    vehicle_id        UUID NOT NULL REFERENCES vehicles(id),
    service_type      TEXT NOT NULL,           -- engine_oil | tire_rotation | brake_pads | belt | full_service | custom
    description       TEXT,
    interval_days     INT,                     -- repeat every N days (NULL = no time-based trigger)
    interval_km       INT,                     -- repeat every N kilometers (NULL = no distance-based trigger)
    last_service_at   TIMESTAMPTZ,
    last_odometer_m   BIGINT,                  -- odometer at last service (meters)
    next_due_at       TIMESTAMPTZ,             -- computed or manually set
    next_due_odometer BIGINT,                  -- next service odometer threshold (meters)
    warn_days_before  INT DEFAULT 7,           -- start warning N days before due date
    warn_km_before    INT DEFAULT 500,         -- start warning N km before due odometer (km)
    enabled           BOOL NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX idx_service_schedules_vehicle ON service_schedules(vehicle_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_service_schedules_tenant  ON service_schedules(tenant_id, enabled) WHERE deleted_at IS NULL;

-- Service log: records of completed maintenance
CREATE TABLE IF NOT EXISTS service_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id),
    schedule_id     UUID REFERENCES service_schedules(id),
    service_type    TEXT NOT NULL,
    description     TEXT,
    serviced_at     TIMESTAMPTZ NOT NULL,
    odometer_m      BIGINT,                    -- odometer reading at service time (meters)
    technician      TEXT,
    service_center  TEXT,
    cost            NUMERIC(12,2),
    currency        TEXT DEFAULT 'INR',
    notes           TEXT,
    parts_used      JSONB DEFAULT '[]',        -- [{"name":"Oil Filter","part_no":"...","qty":1,"cost":150}]
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_service_log_vehicle ON service_log(vehicle_id, serviced_at DESC);
CREATE INDEX idx_service_log_tenant  ON service_log(tenant_id, serviced_at DESC);

-- Vehicle documents with expiry tracking
CREATE TABLE IF NOT EXISTS vehicle_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id),
    doc_type    TEXT NOT NULL,       -- fitness_certificate | insurance | pollution | registration | permit
    doc_number  TEXT,
    issued_at   DATE,
    expires_at  DATE,
    file_url    TEXT,                -- S3 presigned or permanent URL
    issuer      TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_vehicle_documents_vehicle ON vehicle_documents(vehicle_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_vehicle_documents_expiry  ON vehicle_documents(expires_at) WHERE deleted_at IS NULL;

-- Spare parts inventory
CREATE TABLE IF NOT EXISTS spare_parts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    name              TEXT NOT NULL,
    part_number       TEXT,
    description       TEXT,
    qty_in_stock      INT NOT NULL DEFAULT 0,
    reorder_threshold INT NOT NULL DEFAULT 5,
    unit_cost         NUMERIC(12,2),
    currency          TEXT DEFAULT 'INR',
    supplier          TEXT,
    supplier_contact  TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX idx_spare_parts_tenant ON spare_parts(tenant_id) WHERE deleted_at IS NULL;

-- Driver documents
CREATE TABLE IF NOT EXISTS driver_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    driver_id   UUID NOT NULL REFERENCES drivers(id),
    doc_type    TEXT NOT NULL,   -- license | medical_certificate | aadhar | training_certificate
    doc_number  TEXT,
    issued_at   DATE,
    expires_at  DATE,
    file_url    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_driver_documents_driver ON driver_documents(driver_id) WHERE deleted_at IS NULL;

-- Driver score snapshots (weekly)
CREATE TABLE IF NOT EXISTS driver_score_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    driver_id   UUID NOT NULL REFERENCES drivers(id),
    week_start  DATE NOT NULL,
    score       SMALLINT NOT NULL,   -- 0-100
    trips       INT NOT NULL DEFAULT 0,
    distance_m  BIGINT NOT NULL DEFAULT 0,
    harsh_accel INT NOT NULL DEFAULT 0,
    harsh_brake INT NOT NULL DEFAULT 0,
    harsh_corner INT NOT NULL DEFAULT 0,
    overspeed   INT NOT NULL DEFAULT 0,
    idle_min    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (driver_id, week_start)
);
CREATE INDEX idx_driver_score_history_driver ON driver_score_history(driver_id, week_start DESC);

-- RLS
ALTER TABLE service_schedules   ENABLE ROW LEVEL SECURITY;
ALTER TABLE service_log         ENABLE ROW LEVEL SECURITY;
ALTER TABLE vehicle_documents   ENABLE ROW LEVEL SECURITY;
ALTER TABLE spare_parts         ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_documents    ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_score_history ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON service_schedules   USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON service_log         USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON vehicle_documents   USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON spare_parts         USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON driver_documents    USING (tenant_id = current_setting('app.tenant_id')::UUID);
CREATE POLICY tenant_isolation ON driver_score_history USING (tenant_id = current_setting('app.tenant_id')::UUID);
