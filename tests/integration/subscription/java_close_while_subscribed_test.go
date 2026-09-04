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

package subscription

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestSubscriptionJava_ClosingNodeWhileSubscribed_DoesNotRace guards against the Java wrapper's
// subscription-polling goroutine racing with Wrapper.Close. The goroutine keeps calling
// PollSubscriptionNative/CloseSubscriptionNative on w.nodeObj for as long as the subscription's
// context is alive, and this test's context (the shared test-state context) is never cancelled
// mid-test, so the subscription is still actively polling when the node closes at test teardown.
// Close should not be able to delete the JNI global ref while a poll is still in flight or about to
// start. Results is intentionally left unset because this test only cares that closing doesn't
// crash or race, not what the subscription yields.
//
// Run this testwith -race.
func TestSubscriptionJava_ClosingNodeWhileSubscribed_DoesNotRace(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.JavaClientType,
		}),
		Actions: []any{
			&action.SubscriptionRequest{
				Request: `subscription {
					User {
						name
					}
				}`,
			},
		},
	}

	execute(t, test)
}
