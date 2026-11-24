package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

const (
	// maxRichMenuImageBytes is the largest rich menu image the Messaging API accepts
	maxRichMenuImageBytes = 1 << 20
	// maxRichMenuBulkUsers is the largest number of users a bulk link or unlink request may name
	maxRichMenuBulkUsers = 500
)

// rawJSONResponse writes a JSON document as is.
// Rich menu objects are returned this way because the generated response types drop
// the type-specific fields of area actions.
type rawJSONResponse []byte

func (r rawJSONResponse) write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(r)
	return err
}

type getRichMenuRawResponse struct{ rawJSONResponse }

func (r getRichMenuRawResponse) VisitGetRichMenuResponse(w http.ResponseWriter) error {
	return r.write(w)
}

type getRichMenuListRawResponse struct{ rawJSONResponse }

func (r getRichMenuListRawResponse) VisitGetRichMenuListResponse(w http.ResponseWriter) error {
	return r.write(w)
}

// CreateRichMenu creates a rich menu
func (s *server) CreateRichMenu(ctx context.Context, request messagingapi.CreateRichMenuRequestObject) (messagingapi.CreateRichMenuResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}
	if err := validateRichMenu(request.Body); err != nil {
		return nil, err
	}
	content, err := richMenuContent(ctx, request.Body)
	if err != nil {
		return nil, err
	}
	richMenuID, err := newRichMenuID()
	if err != nil {
		return nil, err
	}

	if _, err := db.FromContext(ctx).Queries().CreateRichMenu(ctx, db.CreateRichMenuParams{
		BotID:      auth.GetBotID(ctx),
		RichMenuID: richMenuID,
		Content:    content,
	}); err != nil {
		return nil, fmt.Errorf("failed to create rich menu: %w", err)
	}
	return messagingapi.CreateRichMenu200JSONResponse{RichMenuId: richMenuID}, nil
}

// ValidateRichMenuObject validates a rich menu object without creating it
func (s *server) ValidateRichMenuObject(ctx context.Context, request messagingapi.ValidateRichMenuObjectRequestObject) (messagingapi.ValidateRichMenuObjectResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}
	if err := validateRichMenu(request.Body); err != nil {
		return nil, err
	}
	return messagingapi.ValidateRichMenuObject200Response{}, nil
}

// DeleteRichMenu deletes a rich menu together with its image, aliases and links to users
func (s *server) DeleteRichMenu(ctx context.Context, request messagingapi.DeleteRichMenuRequestObject) (messagingapi.DeleteRichMenuResponseObject, error) {
	deleted, err := db.FromContext(ctx).Queries().DeleteRichMenu(ctx, db.DeleteRichMenuParams{
		BotID:      auth.GetBotID(ctx),
		RichMenuID: request.RichMenuId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete rich menu: %w", err)
	}
	if deleted == 0 {
		return nil, newNotFoundError("richmenu not found")
	}
	return messagingapi.DeleteRichMenu200Response{}, nil
}

// GetRichMenu gets a rich menu
func (s *server) GetRichMenu(ctx context.Context, request messagingapi.GetRichMenuRequestObject) (messagingapi.GetRichMenuResponseObject, error) {
	richMenu, err := getRichMenu(ctx, request.RichMenuId)
	if err != nil {
		return nil, err
	}
	body, err := richMenuJSON(richMenu)
	if err != nil {
		return nil, err
	}
	return getRichMenuRawResponse{body}, nil
}

// GetRichMenuList gets all rich menus of the bot
func (s *server) GetRichMenuList(ctx context.Context, request messagingapi.GetRichMenuListRequestObject) (messagingapi.GetRichMenuListResponseObject, error) {
	richMenus, err := db.FromContext(ctx).Queries().ListRichMenus(ctx, auth.GetBotID(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list rich menus: %w", err)
	}
	items := make([]json.RawMessage, 0, len(richMenus))
	for _, richMenu := range richMenus {
		item, err := richMenuJSON(richMenu)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	body, err := json.Marshal(map[string]any{"richmenus": items})
	if err != nil {
		return nil, fmt.Errorf("failed to encode rich menus: %w", err)
	}
	return getRichMenuListRawResponse{body}, nil
}

// SetRichMenuImage uploads the image of a rich menu. An image can be uploaded only once.
func (s *server) SetRichMenuImage(ctx context.Context, request messagingapi.SetRichMenuImageRequestObject) (messagingapi.SetRichMenuImageResponseObject, error) {
	richMenu, err := getRichMenu(ctx, request.RichMenuId)
	if err != nil {
		return nil, err
	}
	if request.ContentType != "image/jpeg" && request.ContentType != "image/png" {
		return nil, newBadRequestError("invalid content type: only image/jpeg and image/png are accepted")
	}
	image, err := io.ReadAll(io.LimitReader(request.Body, maxRichMenuImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read rich menu image: %w", err)
	}
	if len(image) == 0 {
		return nil, newBadRequestError("the image is empty")
	}
	if len(image) > maxRichMenuImageBytes {
		return nil, newBadRequestError("the image is too large: the maximum size is 1 MB")
	}

	queries := db.FromContext(ctx).Queries()
	hasImage, err := queries.HasRichMenuImage(ctx, richMenu.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check rich menu image: %w", err)
	}
	if hasImage {
		return nil, newBadRequestError("An image has already been uploaded to the richmenu")
	}
	if err := queries.CreateRichMenuImage(ctx, db.CreateRichMenuImageParams{
		RichMenuID:  richMenu.ID,
		Content:     image,
		ContentType: request.ContentType,
	}); err != nil {
		return nil, fmt.Errorf("failed to store rich menu image: %w", err)
	}
	return messagingapi.SetRichMenuImage200Response{}, nil
}

// GetRichMenuImage downloads the image of a rich menu
func (s *server) GetRichMenuImage(ctx context.Context, request messagingapi.GetRichMenuImageRequestObject) (messagingapi.GetRichMenuImageResponseObject, error) {
	richMenu, err := getRichMenu(ctx, request.RichMenuId)
	if err != nil {
		return nil, err
	}
	image, err := db.FromContext(ctx).Queries().GetRichMenuImage(ctx, richMenu.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newNotFoundError("richmenu image not found")
		}
		return nil, fmt.Errorf("failed to get rich menu image: %w", err)
	}
	return messagingapi.GetRichMenuImage200AsteriskResponse{
		Body:          bytes.NewReader(image.Content),
		ContentType:   image.ContentType,
		ContentLength: int64(len(image.Content)),
	}, nil
}

// CreateRichMenuAlias creates a rich menu alias
func (s *server) CreateRichMenuAlias(ctx context.Context, request messagingapi.CreateRichMenuAliasRequestObject) (messagingapi.CreateRichMenuAliasResponseObject, error) {
	if request.Body == nil || request.Body.RichMenuAliasId == "" || request.Body.RichMenuId == "" {
		return messagingapi.CreateRichMenuAlias400JSONResponse{Message: "richMenuAliasId and richMenuId are required"}, nil
	}
	richMenu, err := getRichMenu(ctx, request.Body.RichMenuId)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return messagingapi.CreateRichMenuAlias400JSONResponse{Message: apiErr.Message}, nil
		}
		return nil, err
	}
	created, err := db.FromContext(ctx).Queries().CreateRichMenuAlias(ctx, db.CreateRichMenuAliasParams{
		BotID:      auth.GetBotID(ctx),
		AliasID:    request.Body.RichMenuAliasId,
		RichMenuID: richMenu.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rich menu alias: %w", err)
	}
	if created == 0 {
		return messagingapi.CreateRichMenuAlias400JSONResponse{Message: "conflict richmenu alias id"}, nil
	}
	return messagingapi.CreateRichMenuAlias200Response{}, nil
}

// DeleteRichMenuAlias deletes a rich menu alias
func (s *server) DeleteRichMenuAlias(ctx context.Context, request messagingapi.DeleteRichMenuAliasRequestObject) (messagingapi.DeleteRichMenuAliasResponseObject, error) {
	deleted, err := db.FromContext(ctx).Queries().DeleteRichMenuAlias(ctx, db.DeleteRichMenuAliasParams{
		BotID:   auth.GetBotID(ctx),
		AliasID: request.RichMenuAliasId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete rich menu alias: %w", err)
	}
	if deleted == 0 {
		return nil, newNotFoundError("richmenu alias not found")
	}
	return messagingapi.DeleteRichMenuAlias200Response{}, nil
}

// UpdateRichMenuAlias points a rich menu alias at another rich menu
func (s *server) UpdateRichMenuAlias(ctx context.Context, request messagingapi.UpdateRichMenuAliasRequestObject) (messagingapi.UpdateRichMenuAliasResponseObject, error) {
	if request.Body == nil || request.Body.RichMenuId == "" {
		return messagingapi.UpdateRichMenuAlias400JSONResponse{Message: "richMenuId is required"}, nil
	}
	richMenu, err := getRichMenu(ctx, request.Body.RichMenuId)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return messagingapi.UpdateRichMenuAlias400JSONResponse{Message: apiErr.Message}, nil
		}
		return nil, err
	}
	updated, err := db.FromContext(ctx).Queries().UpdateRichMenuAlias(ctx, db.UpdateRichMenuAliasParams{
		BotID:      auth.GetBotID(ctx),
		AliasID:    request.RichMenuAliasId,
		RichMenuID: richMenu.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update rich menu alias: %w", err)
	}
	if updated == 0 {
		return nil, newNotFoundError("richmenu alias not found")
	}
	return messagingapi.UpdateRichMenuAlias200Response{}, nil
}

// GetRichMenuAlias gets a rich menu alias
func (s *server) GetRichMenuAlias(ctx context.Context, request messagingapi.GetRichMenuAliasRequestObject) (messagingapi.GetRichMenuAliasResponseObject, error) {
	alias, err := db.FromContext(ctx).Queries().GetRichMenuAlias(ctx, db.GetRichMenuAliasParams{
		BotID:   auth.GetBotID(ctx),
		AliasID: request.RichMenuAliasId,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newNotFoundError("richmenu alias not found")
		}
		return nil, fmt.Errorf("failed to get rich menu alias: %w", err)
	}
	return messagingapi.GetRichMenuAlias200JSONResponse{RichMenuAliasId: alias.AliasID, RichMenuId: alias.RichMenuID}, nil
}

// GetRichMenuAliasList gets all rich menu aliases of the bot
func (s *server) GetRichMenuAliasList(ctx context.Context, request messagingapi.GetRichMenuAliasListRequestObject) (messagingapi.GetRichMenuAliasListResponseObject, error) {
	aliases, err := db.FromContext(ctx).Queries().ListRichMenuAliases(ctx, auth.GetBotID(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list rich menu aliases: %w", err)
	}
	resp := messagingapi.GetRichMenuAliasList200JSONResponse{Aliases: make([]messagingapi.RichMenuAliasResponse, 0, len(aliases))}
	for _, alias := range aliases {
		resp.Aliases = append(resp.Aliases, messagingapi.RichMenuAliasResponse{RichMenuAliasId: alias.AliasID, RichMenuId: alias.RichMenuID})
	}
	return resp, nil
}

// SetDefaultRichMenu sets the rich menu shown to users without a per-user rich menu
func (s *server) SetDefaultRichMenu(ctx context.Context, request messagingapi.SetDefaultRichMenuRequestObject) (messagingapi.SetDefaultRichMenuResponseObject, error) {
	richMenu, err := getLinkableRichMenu(ctx, request.RichMenuId)
	if err != nil {
		return nil, err
	}
	if err := db.FromContext(ctx).Queries().UpsertDefaultRichMenu(ctx, db.UpsertDefaultRichMenuParams{
		BotID:      auth.GetBotID(ctx),
		RichMenuID: richMenu.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to set default rich menu: %w", err)
	}
	return messagingapi.SetDefaultRichMenu200Response{}, nil
}

// GetDefaultRichMenuId gets the ID of the default rich menu
func (s *server) GetDefaultRichMenuId(ctx context.Context, request messagingapi.GetDefaultRichMenuIdRequestObject) (messagingapi.GetDefaultRichMenuIdResponseObject, error) {
	richMenu, err := db.FromContext(ctx).Queries().GetDefaultRichMenu(ctx, auth.GetBotID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newNotFoundError("no default richmenu")
		}
		return nil, fmt.Errorf("failed to get default rich menu: %w", err)
	}
	return messagingapi.GetDefaultRichMenuId200JSONResponse{RichMenuId: richMenu.RichMenuID}, nil
}

// CancelDefaultRichMenu removes the default rich menu
func (s *server) CancelDefaultRichMenu(ctx context.Context, request messagingapi.CancelDefaultRichMenuRequestObject) (messagingapi.CancelDefaultRichMenuResponseObject, error) {
	if err := db.FromContext(ctx).Queries().DeleteDefaultRichMenu(ctx, auth.GetBotID(ctx)); err != nil {
		return nil, fmt.Errorf("failed to cancel default rich menu: %w", err)
	}
	return messagingapi.CancelDefaultRichMenu200Response{}, nil
}

// LinkRichMenuIdToUser links a rich menu to a user
func (s *server) LinkRichMenuIdToUser(ctx context.Context, request messagingapi.LinkRichMenuIdToUserRequestObject) (messagingapi.LinkRichMenuIdToUserResponseObject, error) {
	if err := linkRichMenuToUsers(ctx, request.RichMenuId, []string{request.UserId}); err != nil {
		return nil, err
	}
	return messagingapi.LinkRichMenuIdToUser200Response{}, nil
}

// LinkRichMenuIdToUsers links a rich menu to multiple users
func (s *server) LinkRichMenuIdToUsers(ctx context.Context, request messagingapi.LinkRichMenuIdToUsersRequestObject) (messagingapi.LinkRichMenuIdToUsersResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}
	if err := validateBulkUserIDs(request.Body.UserIds); err != nil {
		return nil, err
	}
	if err := linkRichMenuToUsers(ctx, request.Body.RichMenuId, request.Body.UserIds); err != nil {
		return nil, err
	}
	return messagingapi.LinkRichMenuIdToUsers202Response{}, nil
}

// UnlinkRichMenuIdFromUser unlinks the per-user rich menu of a user
func (s *server) UnlinkRichMenuIdFromUser(ctx context.Context, request messagingapi.UnlinkRichMenuIdFromUserRequestObject) (messagingapi.UnlinkRichMenuIdFromUserResponseObject, error) {
	if err := unlinkRichMenuFromUsers(ctx, []string{request.UserId}); err != nil {
		return nil, err
	}
	return messagingapi.UnlinkRichMenuIdFromUser200Response{}, nil
}

// UnlinkRichMenuIdFromUsers unlinks the per-user rich menus of multiple users
func (s *server) UnlinkRichMenuIdFromUsers(ctx context.Context, request messagingapi.UnlinkRichMenuIdFromUsersRequestObject) (messagingapi.UnlinkRichMenuIdFromUsersResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}
	if err := validateBulkUserIDs(request.Body.UserIds); err != nil {
		return nil, err
	}
	if err := unlinkRichMenuFromUsers(ctx, request.Body.UserIds); err != nil {
		return nil, err
	}
	return messagingapi.UnlinkRichMenuIdFromUsers202Response{}, nil
}

// GetRichMenuIdOfUser gets the ID of the per-user rich menu linked to a user.
// The default rich menu isn't reported here.
func (s *server) GetRichMenuIdOfUser(ctx context.Context, request messagingapi.GetRichMenuIdOfUserRequestObject) (messagingapi.GetRichMenuIdOfUserResponseObject, error) {
	richMenu, err := db.FromContext(ctx).Queries().GetUserRichMenu(ctx, db.GetUserRichMenuParams{
		BotID:      auth.GetBotID(ctx),
		LineUserID: request.UserId,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newNotFoundError("the user has no richmenu")
		}
		return nil, fmt.Errorf("failed to get rich menu of user: %w", err)
	}
	return messagingapi.GetRichMenuIdOfUser200JSONResponse{RichMenuId: richMenu.RichMenuID}, nil
}

// richMenuBatchOperation is an operation of a rich menu batch request with its type-specific fields
type richMenuBatchOperation struct {
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

// RichMenuBatch replaces or unlinks the per-user rich menus of many users at once.
// The operations are applied before the response is returned.
func (s *server) RichMenuBatch(ctx context.Context, request messagingapi.RichMenuBatchRequestObject) (messagingapi.RichMenuBatchResponseObject, error) {
	operations, err := richMenuBatchOperations(ctx)
	if err != nil {
		return nil, err
	}
	botID := auth.GetBotID(ctx)
	queries := db.FromContext(ctx).Queries()
	for _, operation := range operations {
		switch operation.Type {
		case "link":
			from, err := getRichMenu(ctx, operation.From)
			if err != nil {
				return nil, err
			}
			to, err := getLinkableRichMenu(ctx, operation.To)
			if err != nil {
				return nil, err
			}
			err = queries.MoveUsersOfRichMenu(ctx, db.MoveUsersOfRichMenuParams{BotID: botID, FromRichMenuID: from.ID, ToRichMenuID: to.ID})
			if err != nil {
				return nil, fmt.Errorf("failed to replace rich menu: %w", err)
			}
		case "unlink":
			from, err := getRichMenu(ctx, operation.From)
			if err != nil {
				return nil, err
			}
			if err := queries.UnlinkUsersOfRichMenu(ctx, db.UnlinkUsersOfRichMenuParams{BotID: botID, RichMenuID: from.ID}); err != nil {
				return nil, fmt.Errorf("failed to unlink rich menu: %w", err)
			}
		case "unlinkAll":
			if err := queries.UnlinkRichMenuFromAllUsers(ctx, botID); err != nil {
				return nil, fmt.Errorf("failed to unlink all rich menus: %w", err)
			}
		default:
			return nil, newBadRequestError(fmt.Sprintf("unknown rich menu batch operation type: %q", operation.Type))
		}
	}
	return messagingapi.RichMenuBatch202Response{}, nil
}

// GetRichMenuBatchProgress reports every batch as succeeded, since batches are applied synchronously
func (s *server) GetRichMenuBatchProgress(ctx context.Context, request messagingapi.GetRichMenuBatchProgressRequestObject) (messagingapi.GetRichMenuBatchProgressResponseObject, error) {
	now := time.Now()
	return messagingapi.GetRichMenuBatchProgress200JSONResponse{
		Phase:         messagingapi.Succeeded,
		AcceptedTime:  now,
		CompletedTime: &now,
	}, nil
}

// ValidateRichMenuBatchRequest validates a rich menu batch request without applying it
func (s *server) ValidateRichMenuBatchRequest(ctx context.Context, request messagingapi.ValidateRichMenuBatchRequestRequestObject) (messagingapi.ValidateRichMenuBatchRequestResponseObject, error) {
	operations, err := richMenuBatchOperations(ctx)
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		if operation.Type != "link" && operation.Type != "unlink" && operation.Type != "unlinkAll" {
			return nil, newBadRequestError(fmt.Sprintf("unknown rich menu batch operation type: %q", operation.Type))
		}
	}
	return messagingapi.ValidateRichMenuBatchRequest200Response{}, nil
}

// richMenuBatchOperations reads the operations from the raw request body,
// because the generated operation type keeps only the type field.
func richMenuBatchOperations(ctx context.Context) ([]richMenuBatchOperation, error) {
	rawBody, ok := rawRequestBody(ctx)
	if !ok {
		return nil, errors.New("raw request body is not available")
	}
	var body struct {
		Operations []richMenuBatchOperation `json:"operations"`
	}
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return nil, newBadRequestError("invalid request body")
	}
	if len(body.Operations) == 0 {
		return nil, newBadRequestError("operations must not be empty")
	}
	return body.Operations, nil
}

// getRichMenu returns a not found APIError when the bot has no such rich menu
func getRichMenu(ctx context.Context, richMenuID string) (db.RichMenu, error) {
	richMenu, err := db.FromContext(ctx).Queries().GetRichMenu(ctx, db.GetRichMenuParams{
		BotID:      auth.GetBotID(ctx),
		RichMenuID: richMenuID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.RichMenu{}, newNotFoundError("richmenu not found")
		}
		return db.RichMenu{}, fmt.Errorf("failed to get rich menu: %w", err)
	}
	return richMenu, nil
}

// getLinkableRichMenu returns a rich menu that can be shown to users, which requires an uploaded image
func getLinkableRichMenu(ctx context.Context, richMenuID string) (db.RichMenu, error) {
	richMenu, err := getRichMenu(ctx, richMenuID)
	if err != nil {
		return db.RichMenu{}, err
	}
	hasImage, err := db.FromContext(ctx).Queries().HasRichMenuImage(ctx, richMenu.ID)
	if err != nil {
		return db.RichMenu{}, fmt.Errorf("failed to check rich menu image: %w", err)
	}
	if !hasImage {
		return db.RichMenu{}, newBadRequestError("must upload richmenu image before applying it to user")
	}
	return richMenu, nil
}

func linkRichMenuToUsers(ctx context.Context, richMenuID string, userIDs []string) error {
	richMenu, err := getLinkableRichMenu(ctx, richMenuID)
	if err != nil {
		return err
	}
	if err := db.FromContext(ctx).Queries().LinkRichMenuToUsers(ctx, db.LinkRichMenuToUsersParams{
		BotID:       auth.GetBotID(ctx),
		RichMenuID:  richMenu.ID,
		LineUserIds: userIDs,
	}); err != nil {
		return fmt.Errorf("failed to link rich menu: %w", err)
	}
	return nil
}

func unlinkRichMenuFromUsers(ctx context.Context, userIDs []string) error {
	if err := db.FromContext(ctx).Queries().UnlinkRichMenuFromUsers(ctx, db.UnlinkRichMenuFromUsersParams{
		BotID:       auth.GetBotID(ctx),
		LineUserIds: userIDs,
	}); err != nil {
		return fmt.Errorf("failed to unlink rich menu: %w", err)
	}
	return nil
}

func validateBulkUserIDs(userIDs []string) error {
	if len(userIDs) == 0 || len(userIDs) > maxRichMenuBulkUsers {
		validationErr := NewValidationError("The request body has 1 error(s)")
		validationErr.AddDetail(fmt.Sprintf("Size must be between 1 and %d", maxRichMenuBulkUsers), "userIds")
		return validationErr
	}
	return nil
}

// validateRichMenu checks the constraints of a rich menu object that don't depend on its area actions
func validateRichMenu(richMenu *messagingapi.RichMenuRequest) error {
	validationErr := NewValidationError("")
	switch {
	case richMenu.Size == nil || richMenu.Size.Width == nil || richMenu.Size.Height == nil:
		validationErr.AddDetail("Must be specified", "size")
	case *richMenu.Size.Width < 800 || *richMenu.Size.Width > 2500:
		validationErr.AddDetail("Must be between 800 and 2500", "size.width")
	case *richMenu.Size.Height < 250:
		validationErr.AddDetail("Must be 250 or more", "size.height")
	case float64(*richMenu.Size.Width)/float64(*richMenu.Size.Height) < 1.45:
		validationErr.AddDetail("The aspect ratio (width / height) must be 1.45 or more", "size")
	}
	if richMenu.Selected == nil {
		validationErr.AddDetail("Must be specified", "selected")
	}
	if richMenu.Name == nil || *richMenu.Name == "" || len([]rune(*richMenu.Name)) > 300 {
		validationErr.AddDetail("Size must be between 1 and 300", "name")
	}
	if richMenu.ChatBarText == nil || *richMenu.ChatBarText == "" || len([]rune(*richMenu.ChatBarText)) > 14 {
		validationErr.AddDetail("Size must be between 1 and 14", "chatBarText")
	}
	if richMenu.Areas == nil || len(*richMenu.Areas) > 20 {
		validationErr.AddDetail("Size must be between 0 and 20", "areas")
	}
	if len(validationErr.Details) == 0 {
		return nil
	}
	validationErr.Message = fmt.Sprintf("The request body has %d error(s)", len(validationErr.Details))
	return validationErr
}

// richMenuContent returns the rich menu object to store. It prefers the raw request body
// so that the type-specific fields of area actions are kept.
func richMenuContent(ctx context.Context, richMenu *messagingapi.RichMenuRequest) ([]byte, error) {
	if rawBody, ok := rawRequestBody(ctx); ok && json.Valid(rawBody) {
		return rawBody, nil
	}
	content, err := json.Marshal(richMenu)
	if err != nil {
		return nil, fmt.Errorf("failed to encode rich menu: %w", err)
	}
	return content, nil
}

// richMenuJSON returns the rich menu object with its ID, as the Messaging API returns it
func richMenuJSON(richMenu db.RichMenu) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(richMenu.Content, &fields); err != nil {
		return nil, fmt.Errorf("failed to decode rich menu %s: %w", richMenu.RichMenuID, err)
	}
	richMenuID, err := json.Marshal(richMenu.RichMenuID)
	if err != nil {
		return nil, fmt.Errorf("failed to encode rich menu ID: %w", err)
	}
	fields["richMenuId"] = richMenuID
	return json.Marshal(fields)
}

func newRichMenuID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", fmt.Errorf("failed to generate rich menu ID: %w", err)
	}
	return "richmenu-" + hex.EncodeToString(id), nil
}
