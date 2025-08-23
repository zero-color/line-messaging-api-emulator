package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/zero-color/line-messaging-api-emulator/db"
)

type botIdContextKey struct {
}

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

			ctx := SetBotID(r.Context(), bot.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetBotID sets the bot ID in the context
// It shouldn't be used outside of this package except for testing
func SetBotID(ctx context.Context, botID int32) context.Context {
	return context.WithValue(ctx, botIdContextKey{}, botID)
}

func GetBotID(ctx context.Context) int32 {
	return ctx.Value(botIdContextKey{}).(int32)
}
