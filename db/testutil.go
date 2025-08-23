package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var testDatabaseURL string

func SetupTestDB() func() {
	pool, err := dockertest.NewPool("")
	pool.MaxWait = 10 * time.Second
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	pwd, _ := os.Getwd()
	dirs := strings.Split(pwd, "/")
	var dbSchemaFilePath string
	for i, dir := range dirs {
		if dir == "line-messaging-api-emulator" {
			dbSchemaFilePath = strings.Join(dirs[:i+1], "/") + "/db/schema.sql"
			break
		}
	}

	runOptions := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17",
		// ポート番号は固定せずに 0 で listen する
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=password",
			"POSTGRES_DB=crm_test",
			"listen_addresses='*'",
		},
		Mounts: []string{
			dbSchemaFilePath + ":/docker-entrypoint-initdb.d/schema.sql",
		},
	}

	resource, err := pool.RunWithOptions(runOptions,
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	time.Sleep(5 * time.Second)
	inspect, err := cli.ContainerInspect(context.Background(), resource.Container.ID)
	if err != nil {
		log.Fatalf("Failed to inspect container: %v", err)
	}
	bindings := inspect.NetworkSettings.Ports["5432/tcp"]
	if len(bindings) == 0 {
		log.Fatal("No port bindings found for 5432/tcp")
	}
	hostPort := bindings[0].HostPort

	hostAndPort := fmt.Sprintf("localhost:%s", hostPort)
	config, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@%s/crm_test?sslmode=disable", hostAndPort))
	if err != nil {
		log.Fatalf("Could not parse config: %s", err)
	}
	databaseURL := stdlib.RegisterConnConfig(config)
	testDatabaseURL = databaseURL

	var db *sql.DB
	if err = pool.Retry(func() error {
		var err error
		db, err = sql.Open("pgx", databaseURL)
		if err != nil {
			return errors.Wrap(err, "sql.Open: %w")
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("connecting to docker failed: %s", err)
	}

	txdb.Register("txdb", "pgx", databaseURL)

	return func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func NewTestDB(t *testing.T) Querier {
	db, err := sql.Open("txdb", uuid.New().String())
	require.NoError(t, err)
	return &Queries{
		db: db,
	}
}
