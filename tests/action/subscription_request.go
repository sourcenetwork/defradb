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

	"github.com/sourcenetwork/defradb/client"
)

// subscriptionTimeout is the maximum time to wait for subscription results to be returned.
const subscriptionTimeout = 1 * time.Second

// SubscriptionRequest represents a subscription request.
//
// The subscription will remain active until shortly after all actions have been processed.
// The results of the subscription will then be asserted upon.
type SubscriptionRequest struct {
	stateful

	// NodeID is the node ID (index) of the node in which to subscribe to.
	NodeID immutable.Option[int]

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

	_, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, node := range nodes {
		result := node.ExecRequest(a.s.Ctx, a.Request)
		if assertErrors(a.s.T, result.GQL.Errors, a.ExpectedError) {
			return
		}

		go func() {
			var results []*client.GQLResult
			for len(results) < len(a.Results) {
				select {
				case s := <-result.Subscription:
					results = append(results, &s)
				case <-time.After(subscriptionTimeout):
				}
			}

			subscriptionAssert <- func() {
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
