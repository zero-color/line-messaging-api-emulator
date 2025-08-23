-- name: CreateBot :one
INSERT INTO bots (
    user_id,
    basic_id,
    chat_mode,
    display_name,
    mark_as_read_mode,
    picture_url,
    premium_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetBot :one
SELECT * FROM bots
WHERE user_id = $1
LIMIT 1;

-- name: GetBotByBasicID :one
SELECT * FROM bots
WHERE basic_id = $1
LIMIT 1;

-- name: ListBots :many
SELECT * FROM bots
ORDER BY created_at DESC;

-- name: UpdateBot :one
UPDATE bots
SET
    basic_id = $2,
    chat_mode = $3,
    display_name = $4,
    mark_as_read_mode = $5,
    picture_url = $6,
    premium_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: DeleteBot :exec
DELETE FROM bots
WHERE user_id = $1;