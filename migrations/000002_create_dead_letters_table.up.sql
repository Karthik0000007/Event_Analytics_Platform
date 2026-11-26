-- One row per message read from the events.dlq Kafka topic.
-- (kafka_partition, kafka_offset) is the position of that message in events.dlq.
-- It is UNIQUE so a replayed DLQ message is stored only once.
CREATE TABLE IF NOT EXISTS dead_letters (
    id              BIGSERIAL PRIMARY KEY,
    original_key    TEXT        NOT NULL DEFAULT '',
    original_value  TEXT        NOT NULL DEFAULT '',
    error_kind      VARCHAR(32) NOT NULL,
    error_message   TEXT        NOT NULL DEFAULT '',
    retry_count     INT         NOT NULL DEFAULT 0,
    failed_at       TIMESTAMPTZ NOT NULL,
    kafka_partition INT         NOT NULL,
    kafka_offset    BIGINT      NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (kafka_partition, kafka_offset)
);

CREATE INDEX IF NOT EXISTS idx_dead_letters_failed_at ON dead_letters (failed_at DESC);
