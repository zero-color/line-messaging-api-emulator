package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/server"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

type options struct {
	Port        uint16 `description:"http port number" long:"port" default:"9090"`
	DatabaseURL string `description:"PostgreSQL connection string" long:"database-url" env:"DATABASE_URL" default:"postgres://postgres:password@localhost:5432/line_emulator?sslmode=disable"`
	SkipSchema  bool   `description:"don't create missing tables and indexes at startup" long:"skip-schema" env:"SKIP_SCHEMA"`
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

	if !opts.SkipSchema {
		if err := db.ApplySchema(context.Background(), sqlDB); err != nil {
			return err
		}
	}

	dbClient := db.NewClient(sqlDB)
	webhookClient := webhook.NewClient()
	r := server.NewRouter(dbClient, server.New(webhookClient))

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
