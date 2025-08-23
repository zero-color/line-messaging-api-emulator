package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		userID = "U" + strings.ReplaceAll(uuid.New().String(), "-", "")
	}

	var basicID string
	if request.Body.BasicId != nil {
		basicID = *request.Body.BasicId
	} else {
		basicID = "@" + shortid.New()
	}

	chatMode := "bot"
	if request.Body.ChatMode != nil {
		chatMode = string(*request.Body.ChatMode)
	}

	markAsReadMode := "manual"
	if request.Body.MarkAsReadMode != nil {
		markAsReadMode = string(*request.Body.MarkAsReadMode)
	}

	var pictureURL pgtype.Text
	if request.Body.PictureUrl != nil {
		pictureURL = pgtype.Text{String: *request.Body.PictureUrl, Valid: true}
	}

	var premiumID pgtype.Text
	if request.Body.PremiumId != nil {
		premiumID = pgtype.Text{String: *request.Body.PremiumId, Valid: true}
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
		BasicId:        bot.BasicID,
		ChatMode:       adminapi.BotInfoResponseChatMode(bot.ChatMode),
		DisplayName:    bot.DisplayName,
		MarkAsReadMode: adminapi.BotInfoResponseMarkAsReadMode(bot.MarkAsReadMode),
		PictureUrl:     pgTextToPtr(bot.PictureUrl),
		PremiumId:      pgTextToPtr(bot.PremiumID),
		UserId:         bot.UserID,
	}

	return adminapi.CreateBot201JSONResponse(response), nil
}

func pgTextToPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

// CreateFollowers creates dummy followers for a bot using bulk insert
func (s *server) CreateFollowers(ctx context.Context, request adminapi.CreateFollowersRequestObject) (adminapi.CreateFollowersResponseObject, error) {
	bot, err := s.db.GetBotByUserID(ctx, request.BotId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminapi.CreateFollowers404JSONResponse{
				Error: struct {
					Code    *string `json:"code,omitempty"`
					Message string  `json:"message"`
				}{
					Message: fmt.Sprintf("Bot with user ID %s not found", request.BotId),
					Code:    lo.ToPtr("NOT_FOUND"),
				},
			}, nil
		}
		return adminapi.CreateFollowers500JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: fmt.Sprintf("Failed to get bot: %v", err),
				Code:    lo.ToPtr("INTERNAL_ERROR"),
			},
		}, nil
	}

	// Validate count
	if err := s.validateFollowerCount(request.Body.Count); err != nil {
		return adminapi.CreateFollowers400JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: err.Error(),
				Code:    lo.ToPtr("INVALID_REQUEST"),
			},
		}, nil
	}

	// Generate dummy user data
	users := s.generateDummyUsers(request.Body.Count)

	// Bulk insert users
	if err := s.bulkInsertUsers(ctx, users); err != nil {
		return adminapi.CreateFollowers500JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: fmt.Sprintf("Failed to create users: %v", err),
				Code:    lo.ToPtr("INTERNAL_ERROR"),
			},
		}, nil
	}

	// Get all created users
	userIDs := extractUserIDs(users)
	createdUsers, err := s.db.GetUsersByUserIDs(ctx, userIDs)
	if err != nil {
		return adminapi.CreateFollowers500JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: fmt.Sprintf("Failed to get created users: %v", err),
				Code:    lo.ToPtr("INTERNAL_ERROR"),
			},
		}, nil
	}

	// Create bot-follower relationships
	if err := s.createBotFollowerRelationships(ctx, bot.ID, createdUsers); err != nil {
		return adminapi.CreateFollowers500JSONResponse{
			Error: struct {
				Code    *string `json:"code,omitempty"`
				Message string  `json:"message"`
			}{
				Message: fmt.Sprintf("Failed to create bot-follower relationships: %v", err),
				Code:    lo.ToPtr("INTERNAL_ERROR"),
			},
		}, nil
	}

	// Build response
	followers := s.buildFollowerProfiles(createdUsers)

	return adminapi.CreateFollowers201JSONResponse{
		Count:     len(followers),
		Followers: followers,
	}, nil
}

// validateFollowerCount ensures the requested count is within acceptable limits
func (s *server) validateFollowerCount(count int) error {
	if count < 1 || count > 1000 {
		return fmt.Errorf("count must be between 1 and 1000")
	}
	return nil
}

// generateDummyUsers creates an array of dummy user data
func (s *server) generateDummyUsers(count int) []db.CreateUsersParams {
	fake := gofakeit.New(0) // Use 0 for random seed based on current time
	users := make([]db.CreateUsersParams, 0, count)

	for i := 0; i < count; i++ {
		user := s.generateSingleUser(fake)
		users = append(users, user)
	}

	return users
}

// generateSingleUser creates a single dummy user with random attributes
func (s *server) generateSingleUser(fake *gofakeit.Faker) db.CreateUsersParams {
	userID := "U" + strings.ReplaceAll(uuid.New().String(), "-", "")

	return db.CreateUsersParams{
		UserID:        userID,
		DisplayName:   generateDisplayName(fake),
		PictureUrl:    generatePictureURL(fake, userID),
		StatusMessage: generateStatusMessage(fake),
		Language:      generateLanguage(fake),
	}
}

// generateDisplayName creates a realistic display name
func generateDisplayName(fake *gofakeit.Faker) string {
	nameType := fake.Number(0, 2)
	switch nameType {
	case 0:
		return fake.Name() // Full name
	case 1:
		return fake.FirstName() // First name only
	case 2:
		return fake.Username() // Username style
	default:
		return fake.Name()
	}
}

// generatePictureURL creates an optional picture URL (70% chance)
func generatePictureURL(fake *gofakeit.Faker, userID string) pgtype.Text {
	if fake.Bool() || fake.Number(1, 10) > 3 {
		return pgtype.Text{
			String: fmt.Sprintf("https://picsum.photos/200?random=%s", userID),
			Valid:  true,
		}
	}
	return pgtype.Text{Valid: false}
}

// generateStatusMessage creates an optional status message (60% chance)
func generateStatusMessage(fake *gofakeit.Faker) pgtype.Text {
	statusMessages := []string{
		"Hello, LINE!",
		"Nice to meet you!",
		"よろしくお願いします",
		"Happy coding!",
		"こんにちは",
		"Living the dream!",
		"Always learning",
		"Coffee ☕",
		"🎵 Music lover",
		"Travel 🌍",
		"", // Some users have no status
	}

	if fake.Number(1, 10) <= 6 {
		if fake.Bool() {
			return pgtype.Text{
				String: fake.RandomString(statusMessages),
				Valid:  true,
			}
		}
		return pgtype.Text{
			String: fake.Sentence(fake.Number(2, 5)),
			Valid:  true,
		}
	}
	return pgtype.Text{Valid: false}
}

// generateLanguage creates an optional language code (80% chance)
func generateLanguage(fake *gofakeit.Faker) pgtype.Text {
	languages := []string{"ja", "en", "zh", "ko", "th", "id", "es", "pt", "fr", "de"}

	if fake.Number(1, 10) <= 8 {
		return pgtype.Text{
			String: fake.RandomString(languages),
			Valid:  true,
		}
	}
	return pgtype.Text{Valid: false}
}

// bulkInsertUsers performs bulk insert of users with fallback to individual inserts
func (s *server) bulkInsertUsers(ctx context.Context, users []db.CreateUsersParams) error {
	_, err := s.db.CreateUsers(ctx, users)
	if err != nil {
		return err
	}
	return nil
}

// extractUserIDs extracts user IDs from the user params array
func extractUserIDs(users []db.CreateUsersParams) []string {
	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.UserID
	}
	return userIDs
}

// createBotFollowerRelationships creates the many-to-many relationships between bot and followers
func (s *server) createBotFollowerRelationships(ctx context.Context, botID int32, users []db.User) error {
	// Prepare bulk data for bot_followers
	followerRecords := make([]db.CreateBotFollowersParams, 0, len(users))
	for _, user := range users {
		followerRecords = append(followerRecords, db.CreateBotFollowersParams{
			BotID:  botID,
			UserID: user.ID,
		})
	}

	// Bulk insert bot-follower relationships
	if _, err := s.db.CreateBotFollowers(ctx, followerRecords); err != nil {
		return err
	}
	return nil
}

// buildFollowerProfiles converts database users to API response format
func (s *server) buildFollowerProfiles(users []db.User) []adminapi.FollowerProfile {
	followers := make([]adminapi.FollowerProfile, 0, len(users))

	for _, user := range users {
		followerProfile := adminapi.FollowerProfile{
			UserId:      user.UserID,
			DisplayName: user.DisplayName,
		}

		if user.PictureUrl.Valid {
			followerProfile.PictureUrl = &user.PictureUrl.String
		}

		if user.StatusMessage.Valid {
			followerProfile.StatusMessage = &user.StatusMessage.String
		}

		if user.Language.Valid {
			followerProfile.Language = &user.Language.String
		}

		followers = append(followers, followerProfile)
	}

	return followers
}
