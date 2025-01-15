CREATE TABLE IF NOT EXISTS outbox_messages (
    message_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_payload JSONB NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS outbox_messages_unsent_created_at ON outbox_messages (created_at) WHERE sent_at IS NULL;
CREATE INDEX IF NOT EXISTS outbox_messages_sent_created_at ON outbox_messages (created_at) WHERE sent_at IS NOT NULL;

CREATE TRIGGER auto_set_outbox_messages_created_at_and_updated_at
    BEFORE INSERT OR UPDATE ON outbox_messages
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_created_at_and_updated_at();

CREATE TABLE IF NOT EXISTS inbox_messages (
    message_uuid UUID PRIMARY KEY,
    message_payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS inbox_messages_unprocessed_created_at ON inbox_messages (created_at) WHERE processed_at IS NULL;
CREATE INDEX IF NOT EXISTS inbox_messages_processed_created_at ON inbox_messages (created_at) WHERE processed_at IS NOT NULL;

CREATE TRIGGER auto_set_inbox_messages_created_at_and_updated_at
    BEFORE INSERT OR UPDATE ON inbox_messages
    FOR EACH ROW
    EXECUTE FUNCTION auto_set_created_at_and_updated_at();
