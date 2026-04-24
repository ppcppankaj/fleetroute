CREATE TYPE maintenance_status AS ENUM ('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'OVERDUE');

CREATE TABLE maintenance_tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    vehicle_id  UUID NOT NULL,
    type        VARCHAR(100) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    status      maintenance_status DEFAULT 'SCHEDULED',
    odometer    DOUBLE PRECISION,
    due_at      TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    cost        NUMERIC(10,2),
    vendor      VARCHAR(255),
    notes       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_maintenance_vehicle ON maintenance_tasks(vehicle_id, due_at);
CREATE INDEX idx_maintenance_tenant  ON maintenance_tasks(tenant_id, status);

CREATE TABLE maintenance_reminders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id     UUID REFERENCES maintenance_tasks(id) ON DELETE CASCADE,
    reminded_at TIMESTAMPTZ DEFAULT NOW()
);
