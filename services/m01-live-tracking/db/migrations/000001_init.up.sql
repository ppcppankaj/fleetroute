-- TimescaleDB or Postgres extension for spatial if available, but staying simple:

CREATE TABLE breadcrumbs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id   UUID NOT NULL,
    tenant_id    UUID NOT NULL,
    lat          DOUBLE PRECISION NOT NULL,
    lng          DOUBLE PRECISION NOT NULL,
    speed        DOUBLE PRECISION DEFAULT 0,
    heading      DOUBLE PRECISION DEFAULT 0,
    altitude     DOUBLE PRECISION DEFAULT 0,
    ignition     BOOLEAN DEFAULT false,
    timestamp    TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- For TimescaleDB we would make it a hypertable, but standard Postgres works too for the prototype
CREATE INDEX idx_breadcrumbs_vehicle_time ON breadcrumbs(vehicle_id, timestamp DESC);
CREATE INDEX idx_breadcrumbs_tenant_time ON breadcrumbs(tenant_id, timestamp DESC);
