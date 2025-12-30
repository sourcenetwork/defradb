// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
