package main

import (
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

	// Use strict handlers with empty middleware for now
	messagingHandler := messagingapi.NewStrictHandler(s, nil)
	messagingapi.HandlerFromMux(messagingHandler, r)

	adminHandler := adminapi.NewStrictHandler(s, nil)
	adminapi.HandlerFromMux(adminHandler, r)

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
