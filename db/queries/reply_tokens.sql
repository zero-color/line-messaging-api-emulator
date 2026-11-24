-- name: CreateReplyTokens :copyfrom
INSERT INTO reply_tokens (
    token,
    bot_id,
    user_id
) VALUES (
    $1, $2, $3
);

-- name: GetReplyToken :one
SELECT * FROM reply_tokens
WHERE bot_id = @bot_id AND token = @token;
