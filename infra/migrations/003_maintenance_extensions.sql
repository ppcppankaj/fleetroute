-- =============================================================================
-- Migration: 003_maintenance_extensions.sql
-- Adds: maintenance_vendors, maintenance_inspections, tyre_management,
--       fuel_logs, fuel_anomalies tables
-- =============================================================================

-- ── Maintenance Vendors ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_vendors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    name            TEXT NOT NULL,
    contact_name    TEXT,
    phone           TEXT,
    email           TEXT,
    address         TEXT,
    services        TEXT[],                  -- e.g., {'oil_change','tyre_service','electrical'}
    rating          NUMERIC(3,1),            -- 1–5 star rating
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_maintenance_vendors_tenant ON maintenance_vendors(tenant_id) WHERE deleted_at IS NULL;

-- ── Maintenance Inspections ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id),
    inspection_type TEXT NOT NULL,           -- routine|safety|emission|annual|pre_trip
    performed_by    TEXT,                    -- Inspector name or vendor
    vendor_id       UUID REFERENCES maintenance_vendors(id),
    inspected_at    TIMESTAMPTZ NOT NULL,
    result          TEXT NOT NULL DEFAULT 'pass',  -- pass|fail|conditional
    notes           TEXT,
    checklist       JSONB NOT NULL DEFAULT '{}',   -- {item: pass|fail|na}
    next_due_at     DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_inspections_vehicle ON maintenance_inspections(vehicle_id, inspected_at DESC);
CREATE INDEX IF NOT EXISTS idx_inspections_tenant  ON maintenance_inspections(tenant_id, inspected_at DESC);

-- ── Tyre Management ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tyre_management (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id),
    position        TEXT NOT NULL,           -- FL|FR|RL|RR|spare|RMI|RMO etc.
    brand           TEXT,
    size            TEXT,                    -- e.g., 275/70R22.5
    serial_number   TEXT,
    fitted_at       DATE,
    fitted_km       BIGINT,                  -- odometer at fitting (metres)
    replaced_at     DATE,
    tread_depth_mm  NUMERIC(5,2),
    condition       TEXT NOT NULL DEFAULT 'good',  -- good|worn|replace
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tyres_vehicle ON tyre_management(vehicle_id, condition);
CREATE INDEX IF NOT EXISTS idx_tyres_tenant  ON tyre_management(tenant_id);

-- ── Fuel Logs ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fuel_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id),
    driver_id       UUID REFERENCES drivers(id),
    liters          NUMERIC(10,3) NOT NULL,
    cost_per_liter  NUMERIC(10,4),
    total_cost      NUMERIC(12,2) GENERATED ALWAYS AS (liters * cost_per_liter) STORED,
    currency        TEXT NOT NULL DEFAULT 'INR',
    odometer_km     BIGINT,                  -- km at fillup
    station_name    TEXT,
    station_lat     DOUBLE PRECISION,
    station_lon     DOUBLE PRECISION,
    fill_type       TEXT NOT NULL DEFAULT 'full',  -- full|partial
    receipt_url     TEXT,
    filled_at       TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_vehicle ON fuel_logs(vehicle_id, filled_at DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_tenant  ON fuel_logs(tenant_id, filled_at DESC);

-- ── Fuel Anomalies ────────────────────────────────────────────────────────────
-- Populated by the stream-processor when it detects rapid fuel level drops
CREATE TABLE IF NOT EXISTS fuel_anomalies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id),
    device_id       UUID NOT NULL,
    anomaly_type    TEXT NOT NULL DEFAULT 'theft',    -- theft|siphon|sensor_fault
    drop_liters     NUMERIC(10,3),
    drop_percent    NUMERIC(5,2),
    start_level     NUMERIC(5,2),
    end_level       NUMERIC(5,2),
    detected_at     TIMESTAMPTZ NOT NULL,
    location        GEOGRAPHY(POINT, 4326),
    confirmed       BOOLEAN NOT NULL DEFAULT FALSE,
    confirmed_by    UUID REFERENCES users(id),
    confirmed_at    TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_fuel_anomalies_vehicle  ON fuel_anomalies(vehicle_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_anomalies_tenant   ON fuel_anomalies(tenant_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_anomalies_unconfirmed ON fuel_anomalies(tenant_id) WHERE NOT confirmed;

-- ── Odometer column on vehicles (in case missing) ─────────────────────────────
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_odometer_m BIGINT DEFAULT 0;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_capacity_liters NUMERIC(7,2);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_type TEXT;  -- petrol|diesel|cng|ev

COMMENT ON TABLE fuel_logs IS 'M09: Manual and sensor-derived fuel fill-up records';
COMMENT ON TABLE fuel_anomalies IS 'M09: AI-detected rapid fuel drop events (theft/siphon)';
COMMENT ON TABLE maintenance_vendors IS 'M08: Service center / workshop records';
COMMENT ON TABLE maintenance_inspections IS 'M08: Vehicle inspection records with checklist';
COMMENT ON TABLE tyre_management IS 'M08: Tyre fitment and condition tracking';
