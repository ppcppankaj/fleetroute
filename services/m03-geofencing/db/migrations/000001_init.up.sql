CREATE TABLE geofences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        VARCHAR(150) NOT NULL,
    type        VARCHAR(50) NOT NULL, -- CIRCLE, POLYGON
    properties  JSONB, -- Radius if circle, metadata
    polygon     JSONB NOT NULL, -- Array of coordinates for simplicty (no postgis requirement for mvp)
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Store current state to emit events ONLY on ENTRY/EXIT
CREATE TABLE vehicle_geofence_states (
    vehicle_id  UUID NOT NULL,
    geofence_id UUID NOT NULL,
    tenant_id   UUID NOT NULL,
    state       VARCHAR(20) NOT NULL, -- INSIDE, OUTSIDE
    last_event  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (vehicle_id, geofence_id)
);
