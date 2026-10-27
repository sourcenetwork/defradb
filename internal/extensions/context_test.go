// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package extensions

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestCollect_WithoutAccumulator_ReturnsNil(t *testing.T) {
	require.Nil(t, Collect(context.Background()))
}

func TestAddWarning_WithoutAccumulator_DoesNothing(t *testing.T) {
	ctx := context.Background()
	AddWarning(ctx, client.GQLWarning{Code: "ignored"})

	require.Nil(t, Collect(ctx))
}

func TestCollect_WithEmptyAccumulator_ReturnsNil(t *testing.T) {
	ctx := WithAccumulator(context.Background())

	require.Nil(t, Collect(ctx))
}

func TestCollect_WithWarnings_PreservesOrder(t *testing.T) {
	ctx := WithAccumulator(context.Background())
	AddWarning(ctx, client.GQLWarning{Code: "first"})
	AddWarning(ctx, client.GQLWarning{Code: "second"})

	result := Collect(ctx)
	require.NotNil(t, result)
	require.Len(t, result.Warnings, 2)
	require.Equal(t, "first", result.Warnings[0].Code)
	require.Equal(t, "second", result.Warnings[1].Code)
}

// A subscription builds a new context for each event, so each event gets its own
// accumulator. If they shared one, the second event would repeat the first's warning.
func TestWithAccumulator_NestedContext_IsolatesWarnings(t *testing.T) {
	first := WithAccumulator(context.Background())
	AddWarning(first, client.GQLWarning{Code: "first"})

	second := WithAccumulator(first)
	AddWarning(second, client.GQLWarning{Code: "second"})

	firstResult := Collect(first)
	require.Len(t, firstResult.Warnings, 1)
	require.Equal(t, "first", firstResult.Warnings[0].Code)

	secondResult := Collect(second)
	require.Len(t, secondResult.Warnings, 1)
	require.Equal(t, "second", secondResult.Warnings[0].Code)
}

func TestAddWarning_FromMultipleGoroutines_RecordsAll(t *testing.T) {
	ctx := WithAccumulator(context.Background())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			AddWarning(ctx, client.GQLWarning{Code: "concurrent"})
		}()
	}
	wg.Wait()

	require.Len(t, Collect(ctx).Warnings, 50)
}

func TestCollect_ThenAddWarning_DoesNotChangeTheEarlierResult(t *testing.T) {
	ctx := WithAccumulator(context.Background())
	AddWarning(ctx, client.GQLWarning{Code: "first"})

	collected := Collect(ctx)
	require.Len(t, collected.Warnings, 1)

	AddWarning(ctx, client.GQLWarning{Code: "second"})

	require.Len(t, collected.Warnings, 1)
	require.Equal(t, "first", collected.Warnings[0].Code)
	require.Len(t, Collect(ctx).Warnings, 2)
}
