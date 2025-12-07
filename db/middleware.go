package db

import "net/http"

// HTTPMiddleware is a middleware to inject DB into the context
func HTTPMiddleware(db DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = ToContext(ctx, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
