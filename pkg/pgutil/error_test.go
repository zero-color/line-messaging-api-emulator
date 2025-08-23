package pgutil

import (
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolationError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				err: nil,
			},
			want: false,
		},
		{
			name: "not unique violation",
			args: args{
				err: errors.New("some error"),
			},
			want: false,
		},
		{
			name: "unique violation",
			args: args{
				err: &pgconn.PgError{
					Code: "23505",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniqueViolationError(tt.args.err); got != tt.want {
				t.Errorf("IsUniqueViolationError() = %v, want %v", got, tt.want)
			}
		})
	}
}
