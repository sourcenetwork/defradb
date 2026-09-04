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

//go:build javaclient

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
)

// TestWrapperSubscription_ContextCancelled_ChannelClosesPromptly guards against
// wrapSubscriptionAsChannel's background poll goroutine outliving the caller's context.
// Cancelling the subscription's context must make the goroutine exit (and so close the returned
// channel) promptly, rather than continuing to poll.
//
// This only checks the process-local effect (the goroutine exits). It can't directly verify from
// this package that CloseSubscriptionNative actually released the C-side subscription store entry
// too, since that store is private to the cbindings package. TestSubscriptionJava_
// ClosingNodeWhileSubscribed_DoesNotRace (located in tests/integration/subscription) covers the
// close-while-subscribed/native-cleanup side under -race instead.
func TestWrapperSubscription_ContextCancelled_ChannelClosesPromptly(t *testing.T) {
	w, ctx := newTestWrapper(t)

	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := w.ExecRequest(subCtx, `subscription { Users { name } }`, options.ExecRequest())
	require.Empty(t, result.GQL.Errors)
	require.NotNil(t, result.Subscription)

	cancel()

	select {
	case _, open := <-result.Subscription:
		require.False(t, open, "expected the subscription channel to close after cancellation, got a value instead")
	case <-time.After(time.Second):
		t.Fatal("subscription channel did not close within a second of its context being cancelled")
	}
}
