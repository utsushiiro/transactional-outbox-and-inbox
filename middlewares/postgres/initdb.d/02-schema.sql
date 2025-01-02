CREATE TABLE IF NOT EXISTS outbox_messages (
    message_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_topic TEXT NOT NULL,
    message_payload JSONB NOT NULL,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
