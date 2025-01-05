SELECT cron.schedule(
    'cleanup_outbox',
    '* * * * *',
    $$
    DELETE FROM outbox_messages
    WHERE created_at < now() - INTERVAL '5 minutes'
      AND sent_at IS NOT NULL;
    $$
);

SELECT cron.schedule(
    'cleanup_inbox',
    '* * * * *',
    $$
    DELETE FROM inbox_messages
    WHERE created_at < now() - INTERVAL '5 minutes'
      AND processed_at IS NOT NULL;
    $$
);
