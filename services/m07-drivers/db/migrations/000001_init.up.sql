CREATE TABLE drivers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    phone           VARCHAR(20) NOT NULL,
    email           VARCHAR(255),
    license_no      VARCHAR(50) UNIQUE NOT NULL,
    license_expiry  DATE NOT NULL,
    group_id        UUID,
    behavior_score  DOUBLE PRECISION DEFAULT 100,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE driver_documents (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id  UUID REFERENCES drivers(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL,
    file_url   VARCHAR(500) NOT NULL,
    expires_at DATE
);

CREATE TABLE coaching_notes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id  UUID REFERENCES drivers(id) ON DELETE CASCADE,
    note       TEXT NOT NULL,
    category   VARCHAR(100),
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE behavior_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id  UUID REFERENCES drivers(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL,
    trip_id    UUID,
    type       VARCHAR(50) NOT NULL,
    severity   VARCHAR(20) NOT NULL,
    points     DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_behavior_driver ON behavior_events(driver_id, created_at DESC);
