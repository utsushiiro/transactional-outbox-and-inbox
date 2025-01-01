CREATE TYPE outbox_message_types AS ENUM ('PENDING', 'PUBLISHED');

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    message_type TEXT NOT NULL,
    message_payload JSONB NOT NULL,
    message_status outbox_message_types NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
