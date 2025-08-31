-- name: CreateMessage :one
INSERT INTO messages (
    bot_id, 
    message_type, 
    recipient_type,
    recipient_id,
    content,
    retry_key,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetMessagesByRetryKey :one
SELECT * FROM messages WHERE retry_key = $1 LIMIT 1;

-- name: GetBotMessages :many
SELECT * FROM messages 
WHERE bot_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountBotMessages :one
SELECT COUNT(*) FROM messages WHERE bot_id = $1;