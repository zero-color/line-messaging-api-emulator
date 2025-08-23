package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/pkg/pgutil"
	"github.com/zero-color/line-messaging-api-emulator/pkg/shortid"
)

// CreateBot creates a new bot
func (s *server) CreateBot(ctx context.Context, request adminapi.CreateBotRequestObject) (adminapi.CreateBotResponseObject, error) {
	var userID string
	if request.Body.UserId != nil && *request.Body.UserId != "" {
		userID = *request.Body.UserId
	} else {
		userID = "U" + uuid.New().String()
	}

	var basicID sql.NullString
	if request.Body.BasicId != nil {
		basicID = sql.NullString{String: *request.Body.BasicId, Valid: true}
	} else {
		basicID = sql.NullString{String: "@" + shortid.New(), Valid: true}
	}

	chatMode := "bot"
	if request.Body.ChatMode != nil {
		chatMode = string(*request.Body.ChatMode)
	}

	markAsReadMode := "manual"
	if request.Body.MarkAsReadMode != nil {
		markAsReadMode = string(*request.Body.MarkAsReadMode)
	}

	var pictureURL sql.NullString
	if request.Body.PictureUrl != nil {
		pictureURL = sql.NullString{String: *request.Body.PictureUrl, Valid: true}
	}

	var premiumID sql.NullString
	if request.Body.PremiumId != nil {
		premiumID = sql.NullString{String: *request.Body.PremiumId, Valid: true}
	}

	bot, err := s.db.CreateBot(ctx, db.CreateBotParams{
		UserID:         userID,
		BasicID:        basicID,
		ChatMode:       chatMode,
		DisplayName:    request.Body.DisplayName,
		MarkAsReadMode: markAsReadMode,
		PictureUrl:     pictureURL,
		PremiumID:      premiumID,
	})

	if err != nil {
		if pgutil.IsUniqueViolationError(err) {
			return adminapi.CreateBot409JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: "Bot already exists with this user ID or basic ID",
					Code:    lo.ToPtr("CONFLICT"),
				},
			}, nil
		}
		return adminapi.CreateBot500JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: fmt.Sprintf("Failed to create bot: %v", err),
				Code:    lo.ToPtr("INTERNAL_ERROR"),
			},
		}, nil
	}

	response := adminapi.BotInfoResponse{
		BasicId:        bot.BasicID.String,
		ChatMode:       adminapi.BotInfoResponseChatMode(bot.ChatMode),
		DisplayName:    bot.DisplayName,
		MarkAsReadMode: adminapi.BotInfoResponseMarkAsReadMode(bot.MarkAsReadMode),
		PictureUrl:     nullStringToPtr(bot.PictureUrl),
		PremiumId:      nullStringToPtr(bot.PremiumID),
		UserId:         bot.UserID,
	}

	return adminapi.CreateBot201JSONResponse(response), nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
