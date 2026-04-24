CREATE TABLE roadmap_features (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    status      VARCHAR(30) DEFAULT 'PLANNED',  -- PLANNED, IN_PROGRESS, DONE, CANCELLED
    category    VARCHAR(100),
    votes       INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE feature_votes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_id  UUID REFERENCES roadmap_features(id) ON DELETE CASCADE,
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(feature_id, user_id)
);
