CREATE TYPE trip_status AS ENUM ('IN_PROGRESS', 'COMPLETED', 'CANCELLED');

CREATE TABLE trips (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    vehicle_id   UUID NOT NULL,
    driver_id    UUID,
    status       trip_status DEFAULT 'IN_PROGRESS',
    start_lat    DOUBLE PRECISION,
    start_lng    DOUBLE PRECISION,
    end_lat      DOUBLE PRECISION,
    end_lng      DOUBLE PRECISION,
    distance_km  DOUBLE PRECISION DEFAULT 0,
    fuel_used    DOUBLE PRECISION DEFAULT 0,
    duration_sec INT DEFAULT 0,
    started_at   TIMESTAMPTZ DEFAULT NOW(),
    ended_at     TIMESTAMPTZ
);
CREATE INDEX idx_trips_vehicle ON trips(vehicle_id, started_at DESC);
CREATE INDEX idx_trips_tenant  ON trips(tenant_id, started_at DESC);

CREATE TABLE trip_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id    UUID REFERENCES trips(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL,
    lat        DOUBLE PRECISION,
    lng        DOUBLE PRECISION,
    speed      DOUBLE PRECISION,
    metadata   JSONB,
    occurred_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_trip_events ON trip_events(trip_id, occurred_at);

CREATE TABLE routes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(150) NOT NULL,
    waypoints   JSONB NOT NULL,
    distance_km DOUBLE PRECISION,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
