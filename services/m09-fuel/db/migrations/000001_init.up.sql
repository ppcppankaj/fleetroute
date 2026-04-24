CREATE TABLE fuel_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id  UUID,
    trip_id    UUID,
    liters     DOUBLE PRECISION NOT NULL,
    total_cost NUMERIC(10,2) NOT NULL,
    price_per_liter DOUBLE PRECISION,
    odometer   DOUBLE PRECISION,
    station    VARCHAR(255),
    fuel_type  VARCHAR(50),
    logged_at  TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_fuel_vehicle ON fuel_logs(vehicle_id, logged_at DESC);
CREATE INDEX idx_fuel_tenant  ON fuel_logs(tenant_id, logged_at DESC);

CREATE TABLE fuel_tank_readings (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    tenant_id  UUID NOT NULL,
    level_pct  DOUBLE PRECISION NOT NULL,
    liters     DOUBLE PRECISION,
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_tank_vehicle ON fuel_tank_readings(vehicle_id, recorded_at DESC);
