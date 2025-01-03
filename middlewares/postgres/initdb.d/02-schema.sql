CREATE TABLE IF NOT EXISTS outbox_messages (
    message_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_topic TEXT NOT NULL,
    message_payload JSONB NOT NULL,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS outbox_messages_unsent_created_at ON outbox_messages (created_at) WHERE sent_at IS NULL;

CREATE TABLE IF NOT EXISTS inbox_messages (
    message_uuid UUID PRIMARY KEY,
    message_payload JSONB NOT NULL,
    received_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS inbox_messages_unprocessed_created_at ON inbox_messages (created_at) WHERE processed_at IS NULL;
