package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/zero-color/line-messaging-api-emulator/db"
)

type contextKey string

const BotIDContextKey contextKey = "bot_id"

func Middleware(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Unauthorized: invalid authorization format", http.StatusUnauthorized)
				return
			}

			botUserID := parts[1]

			bot, err := queries.GetBotByUserID(r.Context(), botUserID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "Unauthorized: bot not found", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), BotIDContextKey, bot.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetBotID(ctx context.Context) int32 {
	return ctx.Value(BotIDContextKey).(int32)
}
