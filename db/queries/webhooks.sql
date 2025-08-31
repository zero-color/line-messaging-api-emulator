-- name: GetWebhook :one
SELECT endpoint, active FROM webhooks
WHERE bot_id = $1;

-- name: UpsertWebhook :exec
INSERT INTO webhooks (bot_id, endpoint, active, updated_at)
VALUES (
        $1,
    $2,
    true,
    CURRENT_TIMESTAMP
)
ON CONFLICT (bot_id) 
DO UPDATE SET 
    endpoint = EXCLUDED.endpoint,
    active = EXCLUDED.active,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetWebhookByBotID :one
SELECT endpoint, active FROM webhooks
WHERE bot_id = $1;