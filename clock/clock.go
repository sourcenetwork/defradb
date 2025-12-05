package clock

import (
	"context"
	"time"
)

type ctxKey struct{}

var clockCtxKey = ctxKey{}

func TimeFromContext(ctx context.Context) time.Time {
	t, ok := ctx.Value(clockCtxKey).(time.Time)
	if !ok {
		return time.Now()
	}

	return t
}

func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, clockCtxKey, t)
}
