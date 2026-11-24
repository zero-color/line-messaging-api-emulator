-- name: CreateRichMenu :one
INSERT INTO rich_menus (bot_id, rich_menu_id, content)
VALUES (@bot_id, @rich_menu_id, @content)
RETURNING *;

-- name: GetRichMenu :one
SELECT * FROM rich_menus
WHERE bot_id = @bot_id AND rich_menu_id = @rich_menu_id;

-- name: ListRichMenus :many
SELECT * FROM rich_menus
WHERE bot_id = @bot_id
ORDER BY id;

-- name: DeleteRichMenu :execrows
DELETE FROM rich_menus
WHERE bot_id = @bot_id AND rich_menu_id = @rich_menu_id;

-- name: CreateRichMenuImage :exec
INSERT INTO rich_menu_images (rich_menu_id, content, content_type)
VALUES (@rich_menu_id, @content, @content_type);

-- name: GetRichMenuImage :one
SELECT * FROM rich_menu_images
WHERE rich_menu_id = @rich_menu_id;

-- name: HasRichMenuImage :one
SELECT EXISTS (
    SELECT 1 FROM rich_menu_images WHERE rich_menu_id = @rich_menu_id
);

-- name: CreateRichMenuAlias :execrows
-- Affects no rows when the bot already has an alias with the ID.
INSERT INTO rich_menu_aliases (bot_id, alias_id, rich_menu_id)
VALUES (@bot_id, @alias_id, @rich_menu_id)
ON CONFLICT (bot_id, alias_id) DO NOTHING;

-- name: UpdateRichMenuAlias :execrows
UPDATE rich_menu_aliases
SET rich_menu_id = @rich_menu_id
WHERE bot_id = @bot_id AND alias_id = @alias_id;

-- name: DeleteRichMenuAlias :execrows
DELETE FROM rich_menu_aliases
WHERE bot_id = @bot_id AND alias_id = @alias_id;

-- name: GetRichMenuAlias :one
SELECT a.alias_id, rm.rich_menu_id FROM rich_menu_aliases a
INNER JOIN rich_menus rm ON rm.id = a.rich_menu_id
WHERE a.bot_id = @bot_id AND a.alias_id = @alias_id;

-- name: ListRichMenuAliases :many
SELECT a.alias_id, rm.rich_menu_id FROM rich_menu_aliases a
INNER JOIN rich_menus rm ON rm.id = a.rich_menu_id
WHERE a.bot_id = @bot_id
ORDER BY a.created_at, a.alias_id;

-- name: UpsertDefaultRichMenu :exec
INSERT INTO default_rich_menus (bot_id, rich_menu_id)
VALUES (@bot_id, @rich_menu_id)
ON CONFLICT (bot_id) DO UPDATE SET rich_menu_id = EXCLUDED.rich_menu_id;

-- name: GetDefaultRichMenu :one
SELECT rm.* FROM default_rich_menus d
INNER JOIN rich_menus rm ON rm.id = d.rich_menu_id
WHERE d.bot_id = @bot_id;

-- name: DeleteDefaultRichMenu :exec
DELETE FROM default_rich_menus
WHERE bot_id = @bot_id;

-- name: LinkRichMenuToUsers :exec
-- User IDs that don't exist are skipped.
INSERT INTO user_rich_menus (bot_id, user_id, rich_menu_id)
SELECT @bot_id, u.id, @rich_menu_id FROM users u
WHERE u.user_id = ANY(@line_user_ids::text[])
ON CONFLICT (bot_id, user_id) DO UPDATE SET rich_menu_id = EXCLUDED.rich_menu_id;

-- name: UnlinkRichMenuFromUsers :exec
DELETE FROM user_rich_menus urm
USING users u
WHERE urm.user_id = u.id
  AND urm.bot_id = @bot_id
  AND u.user_id = ANY(@line_user_ids::text[]);

-- name: UnlinkRichMenuFromAllUsers :exec
DELETE FROM user_rich_menus
WHERE bot_id = @bot_id;

-- name: UnlinkUsersOfRichMenu :exec
DELETE FROM user_rich_menus
WHERE bot_id = @bot_id AND rich_menu_id = @rich_menu_id;

-- name: MoveUsersOfRichMenu :exec
UPDATE user_rich_menus
SET rich_menu_id = @to_rich_menu_id
WHERE bot_id = @bot_id AND rich_menu_id = @from_rich_menu_id;

-- name: GetUserRichMenu :one
SELECT rm.* FROM user_rich_menus urm
INNER JOIN rich_menus rm ON rm.id = urm.rich_menu_id
INNER JOIN users u ON u.id = urm.user_id
WHERE urm.bot_id = @bot_id AND u.user_id = @line_user_id;
