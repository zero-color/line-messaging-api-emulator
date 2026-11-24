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

-- name: ListMessagesDeliveredToUser :many
-- Messages that the bot sent and the user received, oldest first.
-- Broadcast and narrowcast messages count only while the user follows the bot.
SELECT m.* FROM messages m
WHERE m.bot_id = @bot_id
  AND (
    (m.message_type = 'push' AND m.recipient_id = @line_user_id::text)
    OR (m.message_type = 'multicast' AND @line_user_id::text = ANY(string_to_array(m.recipient_id, ',')))
    OR (m.message_type = 'reply' AND EXISTS (
        SELECT 1 FROM reply_tokens rt
        INNER JOIN users u ON u.id = rt.user_id
        WHERE rt.bot_id = m.bot_id AND rt.token = m.recipient_id AND u.user_id = @line_user_id::text
    ))
    OR (m.message_type IN ('broadcast', 'narrowcast') AND EXISTS (
        SELECT 1 FROM bot_followers bf
        INNER JOIN users u ON u.id = bf.user_id
        WHERE bf.bot_id = m.bot_id AND u.user_id = @line_user_id::text
    ))
  )
ORDER BY m.created_at, m.id;
