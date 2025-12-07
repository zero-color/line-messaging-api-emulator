package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	DBTX
	Queries() *Queries
}

type dbClient struct {
	*pgxpool.Pool
}

func NewClient(pool *pgxpool.Pool) *dbClient {
	return &dbClient{
		Pool: pool,
	}
}

func (d *dbClient) Queries() *Queries {
	return New(d)
}

type txDBClient struct {
	pgx.Tx
}

func newTxClient(tx pgx.Tx) *txDBClient {
	return &txDBClient{
		Tx: tx,
	}
}

func (t *txDBClient) Queries() *Queries {
	return New(t)
}
