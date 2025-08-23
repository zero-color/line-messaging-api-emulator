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
		UserId:        user.UserID,
		DisplayName:   user.DisplayName,
		PictureUrl:    user.PictureUrl,
		StatusMessage: user.StatusMessage,
		Language:      user.Language,
	}

	return response, nil
}

// GetFollowers gets follower IDs
func (s *server) GetFollowers(ctx context.Context, request messagingapi.GetFollowersRequestObject) (messagingapi.GetFollowersResponseObject, error) {
	botID := auth.GetBotID(ctx)

	// Default limit is 300 if not specified
	limit := int32(300)
	if request.Params.Limit != nil && *request.Params.Limit > 0 {
		limit = *request.Params.Limit
		// LINE API max limit is 1000
		if limit > 1000 {
			limit = 1000
		}
	}

	// Parse the start token as offset (if provided)
	offset := int32(0)
	if request.Params.Start != nil && *request.Params.Start != "" {
		// Parse start token as integer offset
		var parsedOffset int
		if _, err := fmt.Sscanf(*request.Params.Start, "%d", &parsedOffset); err == nil {
			offset = int32(parsedOffset)
		}
	}

	// Get followers from database
	followers, err := s.db.GetBotFollowers(ctx, db.GetBotFollowersParams{
		BotID:  botID,
		Limit:  limit + 1, // Get one extra to check if there are more
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	// Prepare user IDs array
	userIDs := make([]string, 0, limit)
	hasMore := false
	
	for i, follower := range followers {
		if int32(i) >= limit {
			// We have more followers than the limit
			hasMore = true
			break
		}
		userIDs = append(userIDs, follower.UserID)
	}

	// Prepare response
	response := messagingapi.GetFollowers200JSONResponse{
		UserIds: userIDs,
	}

	// Add next token if there are more followers
	if hasMore {
		nextOffset := offset + limit
		nextToken := fmt.Sprintf("%d", nextOffset)
		response.Next = &nextToken
	}

	return response, nil
}

// IssueLinkToken issues a link token for account linking
func (s *server) IssueLinkToken(ctx context.Context, request messagingapi.IssueLinkTokenRequestObject) (messagingapi.IssueLinkTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
