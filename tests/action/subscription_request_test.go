// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// Every other subscription test trusts that when it declares "expect no
// events" the harness actually checks the channel. If that check is ever
// silently dropped, those tests pass for the wrong reason and the regression
// is invisible. This test pins that behavior: when an event arrives and the
// caller expected none, the collector must report it so the caller can fail.
func TestCollectSubscriptionResults_EmptyExpected_CapturesDeliveredEvent(t *testing.T) {
	sub := make(chan client.GQLResult, 1)
	actionsDone := make(chan struct{})

	// Deliver one event before signalling that actions are done — this
	// mirrors a real mutation triggering a subscription event the test did
	// not expect.
	sub <- client.GQLResult{Data: "unexpected"}
	close(actionsDone)

	got := collectSubscriptionResults(sub, []map[string]any{}, actionsDone)

	require.Len(t, got, 1, "an event delivered while no events were expected must be captured")
}

// Counterpart to the test above: when no event arrives and none was expected,
// the collector must return nothing so the caller's assertion passes.
func TestCollectSubscriptionResults_EmptyExpected_NoEventDelivered(t *testing.T) {
	sub := make(chan client.GQLResult)
	actionsDone := make(chan struct{})
	close(actionsDone)

	got := collectSubscriptionResults(sub, []map[string]any{}, actionsDone)

	require.Empty(t, got, "no events expected, none delivered — collector must return nothing")
}

// nil expected means the test doesn't care about the stream (e.g. the
// close-while-subscribed test). The collector must return immediately without
// blocking on the channel or the actions-done signal.
func TestCollectSubscriptionResults_NilExpected_ReturnsImmediately(t *testing.T) {
	sub := make(chan client.GQLResult)
	actionsDone := make(chan struct{}) // intentionally never closed

	done := make(chan []*client.GQLResult)
	go func() {
		done <- collectSubscriptionResults(sub, nil, actionsDone)
	}()

	select {
	case got := <-done:
		require.Empty(t, got)
	case <-time.After(emptyResultsGrace * 5):
		t.Fatal("collector blocked when expected results were nil")
	}
}
