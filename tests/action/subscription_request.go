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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// subscriptionTimeout is the max wait per event.
//
// Each delivered event triggers a fresh plan run inside handleSubscription so
// the per-subscriber view can be re-evaluated. On ACP-gated collections that
// plan does a DAC check, and under the source-hub backend the check takes time.
const subscriptionTimeout = 5 * time.Second

// postActionsGrace is how long we keep listening after all expected events
// have been collected (or, for the empty-expected case, after all actions
// have run). Subscription delivery is asynchronous, so an event triggered
// by the last action can still be in flight when actions finish — without
// this small wait we'd stop reading too early and miss unexpected extra
// events, letting tests pass that should have failed.
const postActionsGrace = 50 * time.Millisecond

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
				require.Len(
					a.s.T,
					results,
					len(a.Results),
					"subscription yielded a different number of events than expected",
				)
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
//   - len(expected) == 0: the test is asserting nothing arrives, so we have
//     to actually listen — otherwise the assertion is vacuous and the test
//     passes even when a regression delivers events. Listen until all
//     actions have run, then keep listening for a short grace window to
//     catch trailing events.
//   - len(expected) > 0: collect that many events with a per-event timeout,
//     then keep listening for a grace window so over-delivery (more events
//     than expected) fails the assertion instead of being silently dropped.
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
	default:
		for len(results) < len(expected) {
			select {
			case s := <-sub:
				results = append(results, &s)
			case <-time.After(subscriptionTimeout):
				return results
			}
		}
	}

	graceTimer := time.NewTimer(postActionsGrace)
	defer graceTimer.Stop()
	for {
		select {
		case s := <-sub:
			results = append(results, &s)
		case <-graceTimer.C:
			return results
		}
	}
}
