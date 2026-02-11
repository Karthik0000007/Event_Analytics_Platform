CREATE TABLE IF NOT EXISTS events (
    event_id    UUID PRIMARY KEY,
    event_type  VARCHAR(255) NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_type ON events (event_type);
CREATE INDEX idx_events_received_at ON events (received_at);
