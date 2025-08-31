package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jessevdk/go-flags"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/server"
)

type options struct {
	Port        uint16 `description:"http port number" long:"port" default:"9090"`
	DatabaseURL string `description:"PostgreSQL connection string" long:"database-url" env:"DATABASE_URL" default:"postgres://postgres:password@localhost:5432/line_emulator?sslmode=disable"`
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func realMain() error {
	opts, err := parseOptions()
	if err != nil {
		flagsErr := &flags.Error{}
		if ok := errors.As(err, &flagsErr); !ok {
			return fmt.Errorf("failed to parse flags: %w", err)
		}
		if errors.Is(flagsErr.Type, flags.ErrHelp) {
			return nil
		}
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	logger := slog.Default()

	sqlDB, err := db.ConnectDB(opts.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer sqlDB.Close()

	dbClient := db.New(sqlDB)
	s := server.New(dbClient)
	r := chi.NewRouter()

	// Admin API routes (no auth required)
	adminHandler := adminapi.NewStrictHandler(s, nil)
	adminapi.HandlerFromMux(adminHandler, r)

	// Messaging API routes with auth middleware
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(dbClient))
		
		// Custom error handler for validation errors
		errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
			var validationErr *server.ValidationError
			if errors.As(err, &validationErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(validationErr.ToErrorResponse())
				return
			}
			// Default error handling
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		
		messagingHandler := messagingapi.NewStrictHandlerWithOptions(s, nil, messagingapi.StrictHTTPServerOptions{
			ResponseErrorHandlerFunc: errorHandler,
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
		})
		messagingapi.HandlerFromMux(messagingHandler, r)
	})

	logger.Info("Starting server", slog.Int("port", int(opts.Port)))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), r); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func parseOptions() (*options, error) {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}
	return &opts, nil
}
