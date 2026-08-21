// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package extensions collects warnings about a request so they can be returned in the
// `extensions` field of the GQL response.
//
// The accumulator is kept on the context. Warnings come from the planner and the
// fetchers, far below the code that builds the response, and returning them up through
// every layer in between would change a lot of code that does not care about them.
package extensions

import (
	"context"
	"sync"

	"github.com/sourcenetwork/defradb/client"
)

// accumulatorContextKey is the key type for the request accumulator.
type accumulatorContextKey struct{}

// accumulator holds the warnings reported during one request.
//
// More than one goroutine can report, so access is locked.
type accumulator struct {
	mu       sync.Mutex
	warnings []client.GQLWarning
}

// WithAccumulator returns a context carrying a new, empty accumulator.
//
// Call it once per request, and once per subscription event. One accumulator shared
// across a subscription would make each event repeat every warning before it.
func WithAccumulator(ctx context.Context) context.Context {
	return context.WithValue(ctx, accumulatorContextKey{}, &accumulator{})
}

// AddWarning records a warning on the context's accumulator.
//
// It does nothing if there is no accumulator, so a caller on a path that was never
// wired up reports nothing instead of failing.
func AddWarning(ctx context.Context, warning client.GQLWarning) {
	acc, ok := ctx.Value(accumulatorContextKey{}).(*accumulator)
	if !ok {
		return
	}

	acc.mu.Lock()
	defer acc.mu.Unlock()
	acc.warnings = append(acc.warnings, warning)
}

// Collect returns the warnings recorded on the context, or nil if there are none.
func Collect(ctx context.Context) *client.GQLExtensions {
	acc, ok := ctx.Value(accumulatorContextKey{}).(*accumulator)
	if !ok {
		return nil
	}

	acc.mu.Lock()
	defer acc.mu.Unlock()
	if len(acc.warnings) == 0 {
		return nil
	}

	return &client.GQLExtensions{Warnings: acc.warnings}
}
