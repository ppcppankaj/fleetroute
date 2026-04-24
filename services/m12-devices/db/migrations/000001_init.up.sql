CREATE TYPE device_status AS ENUM ('UNPROVISIONED', 'ACTIVE', 'OFFLINE', 'FAULTY');

CREATE TABLE devices (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID,
    imei         VARCHAR(20) UNIQUE NOT NULL,
    sim_number   VARCHAR(20),
    sim_iccid    VARCHAR(30),
    model        VARCHAR(100) NOT NULL,
    firmware_ver VARCHAR(50),
    vehicle_id   UUID,
    status       device_status DEFAULT 'UNPROVISIONED',
    config       JSONB,
    last_seen    TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE device_commands (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id  UUID REFERENCES devices(id) ON DELETE CASCADE,
    command    VARCHAR(50) NOT NULL,
    payload    JSONB,
    status     VARCHAR(20) DEFAULT 'PENDING',
    sent_at    TIMESTAMPTZ,
    acked_at   TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE packet_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID REFERENCES devices(id) ON DELETE CASCADE,
    raw         TEXT NOT NULL,
    parsed      JSONB,
    received_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_packets_device ON packet_logs(device_id, received_at DESC);

CREATE TABLE ota_updates (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_model VARCHAR(100) NOT NULL,
    version      VARCHAR(50) NOT NULL,
    file_url     VARCHAR(500) NOT NULL,
    changelog    TEXT,
    pushed       INT DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);
