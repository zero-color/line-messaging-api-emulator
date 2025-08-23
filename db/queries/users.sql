-- name: CreateUser :one
INSERT INTO users (
    user_id,
    display_name,
    picture_url,
    status_message,
    language
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateUsers :copyfrom
INSERT INTO users (
    user_id,
    display_name,
    picture_url,
    status_message,
    language
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetUser :one
SELECT * FROM users WHERE user_id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUsersByUserIDs :many
SELECT * FROM users WHERE user_id = ANY($1::text[]);

-- name: CreateBotFollower :one
INSERT INTO bot_followers (
    bot_id,
    user_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: CreateBotFollowers :copyfrom
INSERT INTO bot_followers (
    bot_id,
    user_id
) VALUES (
    $1, $2
);

-- name: GetBotFollowers :many
SELECT u.* FROM users u
INNER JOIN bot_followers bf ON u.id = bf.user_id
WHERE bf.bot_id = $1
ORDER BY bf.followed_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBotFollowerCount :one
SELECT COUNT(*) FROM bot_followers WHERE bot_id = $1;

-- name: GetBotFollowerUserIDs :many
SELECT u.user_id FROM users u
INNER JOIN bot_followers bf ON u.id = bf.user_id
WHERE bf.bot_id = $1
ORDER BY bf.followed_at DESC
LIMIT $2 OFFSET $3;

-- name: IsBotFollower :one
SELECT EXISTS (
    SELECT 1 FROM bot_followers bf
    INNER JOIN users u ON u.id = bf.user_id
    WHERE bf.bot_id = $1 AND u.user_id = $2
);