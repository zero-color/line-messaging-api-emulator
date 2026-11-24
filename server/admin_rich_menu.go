package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
)

// GetUserRichMenu gets the rich menu a user sees in the chat with a bot
func (s *server) GetUserRichMenu(ctx context.Context, request adminapi.GetUserRichMenuRequestObject) (adminapi.GetUserRichMenuResponseObject, error) {
	bot, user, err := getBotAndUser(ctx, request.BotId, request.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminapi.GetUserRichMenu404JSONResponse(newAdminError("NOT_FOUND", err.Error())), nil
		}
		return adminapi.GetUserRichMenu500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}

	richMenu, source, err := displayedRichMenu(ctx, bot, user)
	if err != nil {
		return adminapi.GetUserRichMenu500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}
	if source == adminapi.None {
		return adminapi.GetUserRichMenu200JSONResponse{Source: adminapi.None}, nil
	}

	body, err := richMenuJSON(richMenu)
	if err != nil {
		return adminapi.GetUserRichMenu500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}
	var richMenuObject map[string]any
	if err := json.Unmarshal(body, &richMenuObject); err != nil {
		return adminapi.GetUserRichMenu500JSONResponse(newAdminError("INTERNAL_ERROR", fmt.Sprintf("Failed to decode rich menu: %v", err))), nil
	}
	return adminapi.GetUserRichMenu200JSONResponse{Source: source, RichMenu: &richMenuObject}, nil
}

// displayedRichMenu resolves the rich menu shown to the user: the per-user rich menu takes
// priority over the bot's default rich menu.
func displayedRichMenu(ctx context.Context, bot db.Bot, user db.User) (db.RichMenu, adminapi.UserRichMenuResponseSource, error) {
	queries := db.FromContext(ctx).Queries()

	richMenu, err := queries.GetUserRichMenu(ctx, db.GetUserRichMenuParams{BotID: bot.ID, LineUserID: user.UserID})
	if err == nil {
		return richMenu, adminapi.User, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.RichMenu{}, "", fmt.Errorf("failed to get rich menu of user: %w", err)
	}

	richMenu, err = queries.GetDefaultRichMenu(ctx, bot.ID)
	if err == nil {
		return richMenu, adminapi.Default, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.RichMenu{}, "", fmt.Errorf("failed to get default rich menu: %w", err)
	}
	return db.RichMenu{}, adminapi.None, nil
}
