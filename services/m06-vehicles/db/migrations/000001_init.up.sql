CREATE TYPE vehicle_status AS ENUM ('ACTIVE', 'INACTIVE', 'MAINTENANCE');

CREATE TABLE vehicles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    plate_number VARCHAR(20) UNIQUE NOT NULL,
    make         VARCHAR(100) NOT NULL,
    model        VARCHAR(100) NOT NULL,
    year         INT NOT NULL,
    color        VARCHAR(50),
    vin          VARCHAR(17) UNIQUE NOT NULL,
    fuel_type    VARCHAR(30) NOT NULL,
    group_id     UUID,
    status       vehicle_status DEFAULT 'ACTIVE',
    odometer     DOUBLE PRECISION DEFAULT 0,
    photos       TEXT[],
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_vehicles_tenant ON vehicles(tenant_id, status);

CREATE TABLE vehicle_groups (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name      VARCHAR(100) NOT NULL
);

CREATE TABLE vehicle_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id  UUID REFERENCES vehicles(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    file_url    VARCHAR(500) NOT NULL,
    expires_at  DATE,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE vehicle_costs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE CASCADE,
    category   VARCHAR(100) NOT NULL,
    amount     NUMERIC(10,2) NOT NULL,
    date       DATE NOT NULL,
    note       TEXT
);

CREATE TABLE vehicle_assignments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id  UUID UNIQUE REFERENCES vehicles(id),
    driver_id   UUID NOT NULL,
    assigned_at TIMESTAMPTZ DEFAULT NOW()
);
