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
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// subscriptionTimeout is the maximum time to wait for subscription results to be returned.
const subscriptionTimeout = 1 * time.Second

// emptyResultsGrace is how long we keep listening after all actions have run
// when the test expects no events. Subscription delivery is asynchronous, so
// an event triggered by the last action can still be in flight when actions
// finish — without this small wait we'd assert "nothing arrived" before
// giving the pipeline a chance to deliver it.
const emptyResultsGrace = 100 * time.Millisecond

// SubscriptionRequest represents a subscription request.
//
// The subscription will remain active until shortly after all actions have been processed.
// The results of the subscription will then be asserted upon.
type SubscriptionRequest struct {
	stateful

	// NodeID is the node ID (index) of the node in which to subscribe to.
	NodeID immutable.Option[int]

	// The identity of the subscriber. Optional.
	//
	// If an Identity is not provided the subscription can only yield public document(s).
	//
	// If an Identity is provided and the collection has a policy, then the subscription
	// will only yield private document(s) that this Identity is permitted to read.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[state.Identity]

	// The subscription request to submit.
	Request string

	// The expected (data) results yielded through the subscription across its lifetime.
	Results []map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*SubscriptionRequest)(nil)
var _ Stateful = (*SubscriptionRequest)(nil)

// Execute executes the subscription request action.
func (a *SubscriptionRequest) Execute() {
	subscriptionAssert := make(chan func())

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		reqOption := options.ExecRequest()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeIDs[index])
		if identOption.HasValue() {
			reqOption.SetIdentity(identOption.Value())
		}

		result := node.ExecRequest(a.s.Ctx, a.Request, reqOption)
		if assertErrors(a.s.T, result.GQL.Errors, a.ExpectedError) {
			return
		}

		go func() {
			results := collectSubscriptionResults(result.Subscription, a.Results, a.s.AllActionsDone)
			subscriptionAssert <- func() {
				if a.Results == nil {
					return
				}
				if len(a.Results) == 0 {
					require.Empty(
						a.s.T,
						results,
						"subscription yielded events when none were expected",
					)
					return
				}
				for i, r := range a.Results {
					// This assert should be executed from the main test routine
					// so that failures will be properly handled.
					expectedErrorRaised := assertRequestResults(
						a.s,
						results[i],
						r,
						a.ExpectedError,
						nil,
						0,
						true,
					)

					assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
				}
			}
		}()
	}

	a.s.SubscriptionResultsChans = append(a.s.SubscriptionResultsChans, subscriptionAssert)
}

// collectSubscriptionResults reads events from sub and returns them. The
// shape of the wait depends on what the test is asserting:
//
//   - expected == nil: the test doesn't care what the stream delivers (e.g.
//     the close-while-subscribed test). Return immediately with no events.
//   - len(expected) == 0: the test cares — it's asserting nothing arrives —
//     so we have to actually listen, otherwise the assertion is vacuous and
//     the test passes even when a regression delivers events. Listen until
//     all actions have run, then drain for a short grace window to catch
//     events that were in flight when the last action finished. Any event
//     received is returned so the caller can fail the assertion.
//   - len(expected) > 0: collect that many events, each with a per-event
//     timeout.
func collectSubscriptionResults(
	sub <-chan client.GQLResult,
	expected []map[string]any,
	actionsDone <-chan struct{},
) []*client.GQLResult {
	var results []*client.GQLResult
	switch {
	case expected == nil:
		return results
	case len(expected) == 0:
		done := false
		for !done {
			select {
			case s := <-sub:
				results = append(results, &s)
			case <-actionsDone:
				done = true
			}
		}
		graceTimer := time.NewTimer(emptyResultsGrace)
		for {
			select {
			case s := <-sub:
				results = append(results, &s)
			case <-graceTimer.C:
				return results
			}
		}
	default:
		for len(results) < len(expected) {
			select {
			case s := <-sub:
				results = append(results, &s)
			case <-time.After(subscriptionTimeout):
			}
		}
		return results
	}
}
