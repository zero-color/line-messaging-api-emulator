package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

// errInvalidUserAction is returned when the requested action can't be performed in the current state
type errInvalidUserAction struct {
	message string
}

func (e *errInvalidUserAction) Error() string {
	return e.message
}

func newAdminError(code string, message string) adminapi.ErrorResponse {
	var resp adminapi.ErrorResponse
	resp.Error.Code = lo.ToPtr(code)
	resp.Error.Message = message
	return resp
}

// SendUserEvent performs an action as a user and delivers the resulting webhook event to the bot
func (s *server) SendUserEvent(ctx context.Context, request adminapi.SendUserEventRequestObject) (adminapi.SendUserEventResponseObject, error) {
	if request.Body == nil {
		return adminapi.SendUserEvent400JSONResponse(newAdminError("INVALID_REQUEST", "Request body is required")), nil
	}

	bot, user, err := getBotAndUser(ctx, request.BotId, request.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminapi.SendUserEvent404JSONResponse(newAdminError("NOT_FOUND", err.Error())), nil
		}
		return adminapi.SendUserEvent500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}

	event, err := applyUserAction(ctx, bot, user, request.Body)
	if err != nil {
		var invalidAction *errInvalidUserAction
		if errors.As(err, &invalidAction) {
			return adminapi.SendUserEvent400JSONResponse(newAdminError("INVALID_REQUEST", invalidAction.message)), nil
		}
		return adminapi.SendUserEvent500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}

	if err := storeReplyTokens(ctx, bot.ID, map[int32]webhook.Event{user.ID: event}); err != nil {
		return adminapi.SendUserEvent500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}

	resp := adminapi.SendUserEventResponse{
		WebhookEventId: event.ID(),
		Delivered:      true,
	}
	if token := event.ReplyToken(); token != "" {
		resp.ReplyToken = &token
	}
	if err := s.deliverEvent(ctx, bot, event); err != nil {
		resp.Delivered = false
		resp.Detail = lo.ToPtr(err.Error())
	}
	return adminapi.SendUserEvent200JSONResponse(resp), nil
}

// ListUserMessages lists the messages a user received from a bot
func (s *server) ListUserMessages(ctx context.Context, request adminapi.ListUserMessagesRequestObject) (adminapi.ListUserMessagesResponseObject, error) {
	bot, user, err := getBotAndUser(ctx, request.BotId, request.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminapi.ListUserMessages404JSONResponse(newAdminError("NOT_FOUND", err.Error())), nil
		}
		return adminapi.ListUserMessages500JSONResponse(newAdminError("INTERNAL_ERROR", err.Error())), nil
	}

	rows, err := db.FromContext(ctx).Queries().ListMessagesDeliveredToUser(ctx, db.ListMessagesDeliveredToUserParams{
		BotID:      bot.ID,
		LineUserID: user.UserID,
	})
	if err != nil {
		return adminapi.ListUserMessages500JSONResponse(newAdminError("INTERNAL_ERROR", fmt.Sprintf("Failed to list messages: %v", err))), nil
	}

	messages := make([]adminapi.ReceivedMessage, 0, len(rows))
	for _, row := range rows {
		var content []map[string]any
		if err := json.Unmarshal(row.Content, &content); err != nil {
			return adminapi.ListUserMessages500JSONResponse(newAdminError("INTERNAL_ERROR", fmt.Sprintf("Failed to decode message %d: %v", row.ID, err))), nil
		}
		messages = append(messages, adminapi.ReceivedMessage{
			Id:       strconv.Itoa(int(row.ID)),
			SendType: adminapi.ReceivedMessageSendType(row.MessageType),
			SentAt:   row.CreatedAt.Time,
			Messages: content,
		})
	}
	return adminapi.ListUserMessages200JSONResponse{Messages: messages}, nil
}

// getBotAndUser returns an error wrapping pgx.ErrNoRows when either of them doesn't exist
func getBotAndUser(ctx context.Context, botUserID string, userID string) (db.Bot, db.User, error) {
	queries := db.FromContext(ctx).Queries()
	bot, err := queries.GetBotByUserID(ctx, botUserID)
	if err != nil {
		return db.Bot{}, db.User{}, fmt.Errorf("get bot with user ID %s: %w", botUserID, err)
	}
	user, err := queries.GetUser(ctx, userID)
	if err != nil {
		return db.Bot{}, db.User{}, fmt.Errorf("get user with user ID %s: %w", userID, err)
	}
	return bot, user, nil
}

// applyUserAction updates the follow state for the action and builds its webhook event
func applyUserAction(ctx context.Context, bot db.Bot, user db.User, req *adminapi.SendUserEventRequest) (webhook.Event, error) {
	queries := db.FromContext(ctx).Queries()
	now := time.Now()

	switch req.Type {
	case adminapi.Follow:
		inserted, err := queries.UpsertBotFollower(ctx, db.UpsertBotFollowerParams{BotID: bot.ID, UserID: user.ID})
		if err != nil {
			return nil, fmt.Errorf("failed to follow the bot: %w", err)
		}
		if inserted == 0 {
			return nil, &errInvalidUserAction{message: "The user already follows the bot"}
		}
		return webhook.NewFollowEvent(user.UserID, now), nil
	case adminapi.Unfollow:
		deleted, err := queries.DeleteBotFollower(ctx, db.DeleteBotFollowerParams{BotID: bot.ID, UserID: user.ID})
		if err != nil {
			return nil, fmt.Errorf("failed to unfollow the bot: %w", err)
		}
		if deleted == 0 {
			return nil, &errInvalidUserAction{message: "The user doesn't follow the bot"}
		}
		return webhook.NewUnfollowEvent(user.UserID, now), nil
	case adminapi.Message, adminapi.Postback:
		isFollower, err := queries.IsBotFollower(ctx, db.IsBotFollowerParams{BotID: bot.ID, UserID: user.UserID})
		if err != nil {
			return nil, fmt.Errorf("failed to check the follow state: %w", err)
		}
		if !isFollower {
			return nil, &errInvalidUserAction{message: "The user doesn't follow the bot"}
		}
		return buildInteractionEvent(user.UserID, req, now)
	default:
		return nil, &errInvalidUserAction{message: fmt.Sprintf("Unknown event type: %s", req.Type)}
	}
}

func buildInteractionEvent(userID string, req *adminapi.SendUserEventRequest, now time.Time) (webhook.Event, error) {
	if req.Type == adminapi.Message {
		if req.Message == nil || req.Message.Text == "" {
			return nil, &errInvalidUserAction{message: "message.text is required for a message event"}
		}
		return webhook.NewTextMessageEvent(userID, req.Message.Text, now), nil
	}
	if req.Postback == nil || req.Postback.Data == "" {
		return nil, &errInvalidUserAction{message: "postback.data is required for a postback event"}
	}
	return webhook.NewPostbackEvent(userID, req.Postback.Data, lo.FromPtr(req.Postback.Params), now), nil
}

// storeReplyTokens records which user each event's reply token belongs to, keyed by the user's internal ID.
// Events without a reply token are skipped.
func storeReplyTokens(ctx context.Context, botID int32, eventsByUserID map[int32]webhook.Event) error {
	params := make([]db.CreateReplyTokensParams, 0, len(eventsByUserID))
	for userID, event := range eventsByUserID {
		if token := event.ReplyToken(); token != "" {
			params = append(params, db.CreateReplyTokensParams{Token: token, BotID: botID, UserID: userID})
		}
	}
	if len(params) == 0 {
		return nil
	}
	if _, err := db.FromContext(ctx).Queries().CreateReplyTokens(ctx, params); err != nil {
		return fmt.Errorf("failed to store reply tokens: %w", err)
	}
	return nil
}

// deliverEvent sends the event to the bot's webhook endpoint and waits for the response
func (s *server) deliverEvent(ctx context.Context, bot db.Bot, event webhook.Event) error {
	webhookConfig, err := db.FromContext(ctx).Queries().GetWebhookByBotID(ctx, bot.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("webhook endpoint is not set")
		}
		return fmt.Errorf("failed to get webhook: %w", err)
	}
	if !webhookConfig.Active || webhookConfig.Endpoint == "" {
		return errors.New("webhook endpoint is not set")
	}
	return s.webhookClient.SendEvents(ctx, &webhook.SendEventsRequest{
		Endpoint:  webhookConfig.Endpoint,
		BotUserID: bot.UserID,
		Events:    []webhook.Event{event},
	})
}
