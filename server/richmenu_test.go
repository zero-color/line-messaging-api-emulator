package server_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/server"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

// richMenuAPI drives the rich menu endpoints of a test server as one bot
type richMenuAPI struct {
	t       *testing.T
	baseURL string
	bot     db.Bot
}

func newRichMenuObject(name string) map[string]any {
	return map[string]any{
		"size":        map[string]any{"width": 2500, "height": 1686},
		"selected":    false,
		"name":        name,
		"chatBarText": "Tap here",
		"areas": []any{
			map[string]any{
				"bounds": map[string]any{"x": 0, "y": 0, "width": 2500, "height": 1686},
				"action": map[string]any{"type": "postback", "data": "action=buy", "displayText": "Buy"},
			},
		},
	}
}

func (a richMenuAPI) call(method string, path string, body any, out any) int {
	a.t.Helper()
	return callAPI(a.t, a.baseURL, method, path, a.bot.UserID, body, out)
}

func (a richMenuAPI) create(name string) string {
	a.t.Helper()
	var resp struct {
		RichMenuID string `json:"richMenuId"`
	}
	require.Equal(a.t, http.StatusOK, a.call(http.MethodPost, "/v2/bot/richmenu", newRichMenuObject(name), &resp))
	require.NotEmpty(a.t, resp.RichMenuID)
	return resp.RichMenuID
}

func (a richMenuAPI) uploadImage(richMenuID string, contentType string, image []byte) int {
	a.t.Helper()
	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/v2/bot/richmenu/"+richMenuID+"/content", bytes.NewReader(image))
	require.NoError(a.t, err)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+a.bot.UserID)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(a.t, err)
	require.NoError(a.t, resp.Body.Close())
	return resp.StatusCode
}

// createWithImage creates a rich menu that can be shown to users
func (a richMenuAPI) createWithImage(name string) string {
	a.t.Helper()
	richMenuID := a.create(name)
	require.Equal(a.t, http.StatusOK, a.uploadImage(richMenuID, "image/png", []byte("png")))
	return richMenuID
}

// richMenuIDOfUser returns the status and the per-user rich menu ID
func (a richMenuAPI) richMenuIDOfUser(user db.User) (int, string) {
	a.t.Helper()
	var resp struct {
		RichMenuID string `json:"richMenuId"`
	}
	status := a.call(http.MethodGet, "/v2/bot/user/"+user.UserID+"/richmenu", nil, &resp)
	return status, resp.RichMenuID
}

func newRichMenuTestServer(t *testing.T) (context.Context, db.DB, string) {
	t.Helper()
	dbClient := db.NewTestDB(t)
	apiServer := httptest.NewServer(server.NewRouter(dbClient, server.New(webhook.NewMockClient())))
	t.Cleanup(apiServer.Close)
	return db.ToContext(context.Background(), dbClient), dbClient, apiServer.URL
}

func TestRichMenu(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	t.Run("a created rich menu is returned with its area actions intact", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")

		var got map[string]any
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/"+richMenuID, nil, &got))

		want := map[string]any{
			"richMenuId":  richMenuID,
			"size":        map[string]any{"width": float64(2500), "height": float64(1686)},
			"selected":    false,
			"name":        "menu",
			"chatBarText": "Tap here",
			"areas": []any{
				map[string]any{
					"bounds": map[string]any{"x": float64(0), "y": float64(0), "width": float64(2500), "height": float64(1686)},
					"action": map[string]any{"type": "postback", "data": "action=buy", "displayText": "Buy"},
				},
			},
		}
		assert.Equal(t, want, got)
	})

	t.Run("the list has the bot's rich menus in creation order and none of another bot", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		otherAPI := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		first := api.create("first")
		otherAPI.create("other bot")
		second := api.create("second")

		var list struct {
			RichMenus []struct {
				RichMenuID string `json:"richMenuId"`
				Name       string `json:"name"`
			} `json:"richmenus"`
		}
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/list", nil, &list))

		require.Len(t, list.RichMenus, 2)
		assert.Equal(t, []string{first, second}, []string{list.RichMenus[0].RichMenuID, list.RichMenus[1].RichMenuID})
		assert.Equal(t, []string{"first", "second"}, []string{list.RichMenus[0].Name, list.RichMenus[1].Name})
	})

	t.Run("the list is empty for a bot without rich menus", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		var list struct {
			RichMenus []map[string]any `json:"richmenus"`
		}
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/list", nil, &list))

		assert.Empty(t, list.RichMenus)
	})

	t.Run("an invalid rich menu is rejected on create and on validate", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		tooLongChatBarText := newRichMenuObject("menu")
		tooLongChatBarText["chatBarText"] = "123456789012345"
		tooNarrow := newRichMenuObject("menu")
		tooNarrow["size"] = map[string]any{"width": 799, "height": 250}
		missingName := newRichMenuObject("")

		for _, body := range []map[string]any{tooLongChatBarText, tooNarrow, missingName} {
			assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu", body, nil))
			assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/validate", body, nil))
		}
		assert.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/validate", newRichMenuObject("menu"), nil))
	})

	t.Run("a chat bar text of exactly 14 characters is accepted", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		body := newRichMenuObject("menu")
		body["chatBarText"] = "ここをタップしてメニュー表示"

		assert.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu", body, nil))
	})

	t.Run("a deleted rich menu is gone and deleting it again is not found", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")

		assert.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/richmenu/"+richMenuID, nil, nil))
		assert.Equal(t, http.StatusNotFound, api.call(http.MethodGet, "/v2/bot/richmenu/"+richMenuID, nil, nil))
		assert.Equal(t, http.StatusNotFound, api.call(http.MethodDelete, "/v2/bot/richmenu/"+richMenuID, nil, nil))
	})

	t.Run("a rich menu of another bot is not found", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		otherAPI := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := otherAPI.create("menu")

		assert.Equal(t, http.StatusNotFound, api.call(http.MethodGet, "/v2/bot/richmenu/"+richMenuID, nil, nil))
	})
}

func TestRichMenuImage(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	download := func(t *testing.T, api richMenuAPI, richMenuID string) (int, string, []byte) {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, baseURL+"/v2/bot/richmenu/"+richMenuID+"/content", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+api.bot.UserID)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, resp.Header.Get("Content-Type"), body
	}

	t.Run("an uploaded image can be downloaded", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")

		require.Equal(t, http.StatusOK, api.uploadImage(richMenuID, "image/jpeg", []byte("jpeg bytes")))

		status, contentType, body := download(t, api, richMenuID)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, "image/jpeg", contentType)
		assert.Equal(t, []byte("jpeg bytes"), body)
	})

	t.Run("a second upload to the same rich menu is rejected", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")
		require.Equal(t, http.StatusOK, api.uploadImage(richMenuID, "image/png", []byte("first")))

		assert.Equal(t, http.StatusBadRequest, api.uploadImage(richMenuID, "image/png", []byte("second")))

		_, _, body := download(t, api, richMenuID)
		assert.Equal(t, []byte("first"), body)
	})

	t.Run("an image of exactly 1 MB is accepted and one byte more is rejected", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		assert.Equal(t, http.StatusBadRequest, api.uploadImage(api.create("too large"), "image/png", bytes.Repeat([]byte{1}, 1<<20+1)))
		assert.Equal(t, http.StatusOK, api.uploadImage(api.create("at the limit"), "image/png", bytes.Repeat([]byte{1}, 1<<20)))
	})

	t.Run("an empty image and an unsupported content type are rejected", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")

		assert.Equal(t, http.StatusBadRequest, api.uploadImage(richMenuID, "image/png", nil))
		assert.Equal(t, http.StatusBadRequest, api.uploadImage(richMenuID, "image/gif", []byte("gif")))
	})

	t.Run("an upload to an unknown rich menu and a download without an image are not found", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		assert.Equal(t, http.StatusNotFound, api.uploadImage("richmenu-unknown", "image/png", []byte("png")))
		status, _, _ := download(t, api, api.create("menu"))
		assert.Equal(t, http.StatusNotFound, status)
	})
}

func TestRichMenuLink(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	t.Run("a bulk link and unlink change the per-user rich menu of the named users only", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		richMenuID := api.createWithImage("menu")
		linked := createTestUser(ctx, t, dbClient, &bot)
		alsoLinked := createTestUser(ctx, t, dbClient, &bot)
		notLinked := createTestUser(ctx, t, dbClient, &bot)

		status := api.call(http.MethodPost, "/v2/bot/richmenu/bulk/link", map[string]any{"richMenuId": richMenuID, "userIds": []string{linked.UserID, alsoLinked.UserID, "Uunknown"}}, nil)
		require.Equal(t, http.StatusAccepted, status)

		status, got := api.richMenuIDOfUser(linked)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, richMenuID, got)
		status, _ = api.richMenuIDOfUser(notLinked)
		assert.Equal(t, http.StatusNotFound, status)

		status = api.call(http.MethodPost, "/v2/bot/richmenu/bulk/unlink", map[string]any{"userIds": []string{linked.UserID}}, nil)
		require.Equal(t, http.StatusAccepted, status)

		status, _ = api.richMenuIDOfUser(linked)
		assert.Equal(t, http.StatusNotFound, status)
		status, got = api.richMenuIDOfUser(alsoLinked)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, richMenuID, got)
	})

	t.Run("linking another rich menu replaces the previous one", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		first := api.createWithImage("first")
		second := api.createWithImage("second")
		user := createTestUser(ctx, t, dbClient, &bot)

		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+first, nil, nil))
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+second, nil, nil))

		_, got := api.richMenuIDOfUser(user)
		assert.Equal(t, second, got)

		require.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/user/"+user.UserID+"/richmenu", nil, nil))
		status, _ := api.richMenuIDOfUser(user)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("a rich menu without an image can't be linked or set as default", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		richMenuID := api.create("no image")
		user := createTestUser(ctx, t, dbClient, &bot)

		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/bulk/link", map[string]any{"richMenuId": richMenuID, "userIds": []string{user.UserID}}, nil))
		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+richMenuID, nil, nil))
		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/"+richMenuID, nil, nil))
	})

	t.Run("a bulk request accepts 500 users and rejects 501 and none", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.createWithImage("menu")
		userIDs := make([]string, 0, 501)
		for i := range 501 {
			userIDs = append(userIDs, fmt.Sprintf("U%032d", i))
		}

		for _, path := range []string{"/v2/bot/richmenu/bulk/link", "/v2/bot/richmenu/bulk/unlink"} {
			assert.Equal(t, http.StatusAccepted, api.call(http.MethodPost, path, map[string]any{"richMenuId": richMenuID, "userIds": userIDs[:500]}, nil), path)
			assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, path, map[string]any{"richMenuId": richMenuID, "userIds": userIDs}, nil), path)
			assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, path, map[string]any{"richMenuId": richMenuID, "userIds": []string{}}, nil), path)
		}
	})

	t.Run("linking an unknown rich menu is not found", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		user := createTestUser(ctx, t, dbClient, &bot)

		assert.Equal(t, http.StatusNotFound, api.call(http.MethodPost, "/v2/bot/richmenu/bulk/link", map[string]any{"richMenuId": "richmenu-unknown", "userIds": []string{user.UserID}}, nil))
	})

	t.Run("deleting a rich menu unlinks it from its users", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		richMenuID := api.createWithImage("menu")
		user := createTestUser(ctx, t, dbClient, &bot)
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+richMenuID, nil, nil))

		require.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/richmenu/"+richMenuID, nil, nil))

		status, _ := api.richMenuIDOfUser(user)
		assert.Equal(t, http.StatusNotFound, status)
	})
}

func TestDefaultRichMenu(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	defaultRichMenuID := func(t *testing.T, api richMenuAPI) (int, string) {
		t.Helper()
		var resp struct {
			RichMenuID string `json:"richMenuId"`
		}
		status := api.call(http.MethodGet, "/v2/bot/user/all/richmenu", nil, &resp)
		return status, resp.RichMenuID
	}

	t.Run("a bot has no default rich menu until one is set, and none again after cancelling", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		first := api.createWithImage("first")
		second := api.createWithImage("second")

		status, _ := defaultRichMenuID(t, api)
		assert.Equal(t, http.StatusNotFound, status)

		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/"+first, nil, nil))
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/"+second, nil, nil))
		status, got := defaultRichMenuID(t, api)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, second, got)

		require.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/user/all/richmenu", nil, nil))
		status, _ = defaultRichMenuID(t, api)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("cancelling without a default rich menu succeeds", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		assert.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/user/all/richmenu", nil, nil))
	})

	t.Run("the default rich menu isn't reported as a user's rich menu", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		user := createTestUser(ctx, t, dbClient, &bot)
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/"+api.createWithImage("menu"), nil, nil))

		status, _ := api.richMenuIDOfUser(user)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("setting an unknown rich menu as default is not found", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		assert.Equal(t, http.StatusNotFound, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/richmenu-unknown", nil, nil))
	})
}

func TestRichMenuAlias(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	type alias struct {
		RichMenuAliasID string `json:"richMenuAliasId"`
		RichMenuID      string `json:"richMenuId"`
	}

	t.Run("an alias can be created, pointed at another rich menu, listed and deleted", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		first := api.create("first")
		second := api.create("second")

		assert.Equal(t, http.StatusNotFound, api.call(http.MethodGet, "/v2/bot/richmenu/alias/tab-a", nil, nil))

		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-a", RichMenuID: first}, nil))
		var got alias
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/alias/tab-a", nil, &got))
		assert.Equal(t, alias{RichMenuAliasID: "tab-a", RichMenuID: first}, got)

		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/alias/tab-a", map[string]any{"richMenuId": second}, nil))
		var list struct {
			Aliases []alias `json:"aliases"`
		}
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/alias/list", nil, &list))
		assert.Equal(t, []alias{{RichMenuAliasID: "tab-a", RichMenuID: second}}, list.Aliases)

		require.Equal(t, http.StatusOK, api.call(http.MethodDelete, "/v2/bot/richmenu/alias/tab-a", nil, nil))
		assert.Equal(t, http.StatusNotFound, api.call(http.MethodGet, "/v2/bot/richmenu/alias/tab-a", nil, nil))
	})

	t.Run("an alias ID can't be used twice by one bot but can be reused by another bot", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		otherAPI := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-a", RichMenuID: richMenuID}, nil))

		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-a", RichMenuID: richMenuID}, nil))
		assert.Equal(t, http.StatusOK, otherAPI.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-a", RichMenuID: otherAPI.create("menu")}, nil))
	})

	t.Run("an alias for an unknown rich menu is rejected", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-a", RichMenuID: api.create("menu")}, nil))

		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/alias", alias{RichMenuAliasID: "tab-b", RichMenuID: "richmenu-unknown"}, nil))
		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/alias/tab-a", map[string]any{"richMenuId": "richmenu-unknown"}, nil))
	})

	t.Run("updating or deleting an unknown alias is not found", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.create("menu")

		assert.Equal(t, http.StatusNotFound, api.call(http.MethodPost, "/v2/bot/richmenu/alias/unknown", map[string]any{"richMenuId": richMenuID}, nil))
		assert.Equal(t, http.StatusNotFound, api.call(http.MethodDelete, "/v2/bot/richmenu/alias/unknown", nil, nil))
	})
}

func TestRichMenuBatch(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	t.Run("link moves the users of one rich menu to another and unlink removes them", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		from := api.createWithImage("from")
		to := api.createWithImage("to")
		untouched := api.createWithImage("untouched")
		movedUser := createTestUser(ctx, t, dbClient, &bot)
		untouchedUser := createTestUser(ctx, t, dbClient, &bot)
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+movedUser.UserID+"/richmenu/"+from, nil, nil))
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+untouchedUser.UserID+"/richmenu/"+untouched, nil, nil))

		operations := map[string]any{"operations": []any{map[string]any{"type": "link", "from": from, "to": to}}}
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/richmenu/validate/batch", operations, nil))
		require.Equal(t, http.StatusAccepted, api.call(http.MethodPost, "/v2/bot/richmenu/batch", operations, nil))

		_, got := api.richMenuIDOfUser(movedUser)
		assert.Equal(t, to, got)
		_, got = api.richMenuIDOfUser(untouchedUser)
		assert.Equal(t, untouched, got)

		require.Equal(t, http.StatusAccepted, api.call(http.MethodPost, "/v2/bot/richmenu/batch", map[string]any{"operations": []any{map[string]any{"type": "unlink", "from": to}}}, nil))
		status, _ := api.richMenuIDOfUser(movedUser)
		assert.Equal(t, http.StatusNotFound, status)
		_, got = api.richMenuIDOfUser(untouchedUser)
		assert.Equal(t, untouched, got)
	})

	t.Run("unlinkAll removes every per-user rich menu of the bot only", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		otherBot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		otherAPI := richMenuAPI{t: t, baseURL: baseURL, bot: otherBot}
		user := createTestUser(ctx, t, dbClient, &bot)
		otherBotMenu := otherAPI.createWithImage("other bot")
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+api.createWithImage("menu"), nil, nil))
		require.Equal(t, http.StatusOK, otherAPI.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+otherBotMenu, nil, nil))

		require.Equal(t, http.StatusAccepted, api.call(http.MethodPost, "/v2/bot/richmenu/batch", map[string]any{"operations": []any{map[string]any{"type": "unlinkAll"}}}, nil))

		status, _ := api.richMenuIDOfUser(user)
		assert.Equal(t, http.StatusNotFound, status)
		_, got := otherAPI.richMenuIDOfUser(user)
		assert.Equal(t, otherBotMenu, got)
	})

	t.Run("no operations, an unknown operation type and an unknown rich menu are rejected", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}
		richMenuID := api.createWithImage("menu")

		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/batch", map[string]any{"operations": []any{}}, nil))
		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/batch", map[string]any{"operations": []any{map[string]any{"type": "swap"}}}, nil))
		assert.Equal(t, http.StatusBadRequest, api.call(http.MethodPost, "/v2/bot/richmenu/validate/batch", map[string]any{"operations": []any{map[string]any{"type": "swap"}}}, nil))
		assert.Equal(t, http.StatusNotFound, api.call(http.MethodPost, "/v2/bot/richmenu/batch", map[string]any{"operations": []any{map[string]any{"type": "link", "from": "richmenu-unknown", "to": richMenuID}}}, nil))
	})

	t.Run("the progress of a batch is reported as succeeded", func(t *testing.T) {
		api := richMenuAPI{t: t, baseURL: baseURL, bot: createTestBot(ctx, t, dbClient, false)}

		var progress struct {
			Phase string `json:"phase"`
		}
		require.Equal(t, http.StatusOK, api.call(http.MethodGet, "/v2/bot/richmenu/progress/batch?requestId=any", nil, &progress))

		assert.Equal(t, "succeeded", progress.Phase)
	})
}

func TestGetUserRichMenu(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	type displayed struct {
		Source   string `json:"source"`
		RichMenu *struct {
			RichMenuID string `json:"richMenuId"`
			Name       string `json:"name"`
			Areas      []struct {
				Action map[string]any `json:"action"`
			} `json:"areas"`
		} `json:"richMenu"`
	}
	getDisplayed := func(t *testing.T, bot db.Bot, user db.User) displayed {
		t.Helper()
		var resp displayed
		status := callAPI(t, baseURL, http.MethodGet, "/admin/bots/"+bot.UserID+"/users/"+user.UserID+"/rich-menu", "", nil, &resp)
		require.Equal(t, http.StatusOK, status)
		return resp
	}

	t.Run("a user sees no rich menu when none is linked and the bot has no default", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		user := createTestUser(ctx, t, dbClient, &bot)

		assert.Equal(t, displayed{Source: "none"}, getDisplayed(t, bot, user))
	})

	t.Run("a user sees the default rich menu until a per-user rich menu is linked", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		api := richMenuAPI{t: t, baseURL: baseURL, bot: bot}
		user := createTestUser(ctx, t, dbClient, &bot)
		otherUser := createTestUser(ctx, t, dbClient, &bot)
		defaultMenu := api.createWithImage("default")
		userMenu := api.createWithImage("per-user")
		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/all/richmenu/"+defaultMenu, nil, nil))

		got := getDisplayed(t, bot, user)
		require.NotNil(t, got.RichMenu)
		assert.Equal(t, "default", got.Source)
		assert.Equal(t, defaultMenu, got.RichMenu.RichMenuID)
		assert.Equal(t, "default", got.RichMenu.Name)
		require.Len(t, got.RichMenu.Areas, 1)
		assert.Equal(t, map[string]any{"type": "postback", "data": "action=buy", "displayText": "Buy"}, got.RichMenu.Areas[0].Action)

		require.Equal(t, http.StatusOK, api.call(http.MethodPost, "/v2/bot/user/"+user.UserID+"/richmenu/"+userMenu, nil, nil))

		got = getDisplayed(t, bot, user)
		require.NotNil(t, got.RichMenu)
		assert.Equal(t, "user", got.Source)
		assert.Equal(t, userMenu, got.RichMenu.RichMenuID)

		other := getDisplayed(t, bot, otherUser)
		require.NotNil(t, other.RichMenu)
		assert.Equal(t, "default", other.Source)
		assert.Equal(t, defaultMenu, other.RichMenu.RichMenuID)
	})

	t.Run("unknown bot or user returns not found", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		user := createTestUser(ctx, t, dbClient, nil)

		assert.Equal(t, http.StatusNotFound, callAPI(t, baseURL, http.MethodGet, "/admin/bots/Uunknown/users/"+user.UserID+"/rich-menu", "", nil, nil))
		assert.Equal(t, http.StatusNotFound, callAPI(t, baseURL, http.MethodGet, "/admin/bots/"+bot.UserID+"/users/Uunknown/rich-menu", "", nil, nil))
	})
}
