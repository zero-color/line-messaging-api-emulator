package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// NewRouter builds the HTTP handler that serves the admin API without authentication
// and the Messaging API behind channel access token authentication.
func NewRouter(dbClient db.DB, s Server) http.Handler {
	r := chi.NewRouter()
	r.Use(db.HTTPMiddleware(dbClient))

	adminHandler := adminapi.NewStrictHandler(s, nil)
	adminapi.HandlerFromMux(adminHandler, r)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware())
		r.Use(RawBodyMiddleware())
		r.Use(dropEmptyRetryKey)

		messagingHandler := messagingapi.NewStrictHandlerWithOptions(s, nil, messagingapi.StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(validationErr.ToErrorResponse())
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			},
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
		})
		messagingapi.HandlerFromMux(messagingHandler, r)
	})
	return r
}

// dropEmptyRetryKey removes an empty X-Line-Retry-Key header, which the LINE Platform treats as absent.
// The official SDKs send the header even when no retry key is given, and the generated
// parameter binding would otherwise reject it as an invalid UUID.
func dropEmptyRetryKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if values, ok := r.Header["X-Line-Retry-Key"]; ok && len(values) == 1 && values[0] == "" {
			r.Header.Del("X-Line-Retry-Key")
		}
		next.ServeHTTP(w, r)
	})
}
