package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// GetProfile gets user profile information
func (s *server) GetProfile(ctx context.Context, request messagingapi.GetProfileRequestObject) (messagingapi.GetProfileResponseObject, error) {
	botID := auth.GetBotID(ctx)

	user, err := s.db.GetBotFollowerUser(ctx, db.GetBotFollowerUserParams{
		BotID:  botID,
		UserID: request.UserId,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := messagingapi.GetProfile200JSONResponse{
		UserId:      user.UserID,
		DisplayName: user.DisplayName,
	}

	if user.PictureUrl.Valid {
		response.PictureUrl = &user.PictureUrl.String
	}

	if user.StatusMessage.Valid {
		response.StatusMessage = &user.StatusMessage.String
	}

	if user.Language.Valid {
		response.Language = &user.Language.String
	}

	return response, nil
}

// GetFollowers gets follower IDs
func (s *server) GetFollowers(ctx context.Context, request messagingapi.GetFollowersRequestObject) (messagingapi.GetFollowersResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// IssueLinkToken issues a link token for account linking
func (s *server) IssueLinkToken(ctx context.Context, request messagingapi.IssueLinkTokenRequestObject) (messagingapi.IssueLinkTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
