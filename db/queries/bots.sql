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
          @user_id,
            @basic_id,
            @chat_mode,
            @display_name,
          @mark_as_read_mode,
            @picture_url,
            @premium_id
) RETURNING *;

-- name: GetBot :one
SELECT * FROM bots
WHERE id = @id;

-- name: GetBotByUserID :one
SELECT * FROM bots
WHERE user_id = @user_id;

-- name: GetBotByBasicID :one
SELECT * FROM bots
WHERE basic_id = @basic_id;

-- name: ListBots :many
SELECT * FROM bots
ORDER BY created_at DESC;

-- name: UpdateBot :one
UPDATE bots
SET
    basic_id = @basic_id,
    chat_mode = @chat_mode,
    display_name = @display_name,
    mark_as_read_mode = @mark_as_read_mode,
    picture_url = @picture_url,
    premium_id = @premium_id,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = @user_id
RETURNING *;

-- name: DeleteBot :exec
DELETE FROM bots
WHERE user_id = @user_id;