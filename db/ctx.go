package db

import "context"

type dbCtxKey struct{}

func ToContext(ctx context.Context, db DB) context.Context {
	return context.WithValue(ctx, dbCtxKey{}, db)
}

func FromContext(ctx context.Context) DB {
	return ctx.Value(dbCtxKey{}).(DB)
}
