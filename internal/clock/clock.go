package clock

import (
	"context"
	"time"

	"github.com/sourcenetwork/defradb/errors"
)

// type Clock = quartz.Clock

// func NewMock()

type ctxKey struct{}

var clockCtxKey = ctxKey{}

func FromContext(ctx context.Context) (time.Time, error) {
	t, ok := ctx.Value(clockCtxKey).(time.Time)
	if !ok {
		return time.Time{}, errors.New("bad clock type")
	}

	return t, nil
}

func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, clockCtxKey, t)
}
