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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// ResultAsserter is an interface that can be implemented to provide custom result
// assertions.
type ResultAsserter interface {
	// Assert will be called with the test and the result of the request.
	Assert(t testing.TB, result map[string]any)
}

// ResultAsserterFunc is a function that can be used to implement the ResultAsserter
type ResultAsserterFunc func(testing.TB, map[string]any) (bool, string)

func (f ResultAsserterFunc) Assert(t testing.TB, result map[string]any) {
	f(t, result)
}

// Request represents a standard Defra (GQL) request.
type Request struct {
	stateful

	// NodeID may hold the ID (index) of a node to execute this request on.
	//
	// If a value is not provided the request will be executed against all nodes,
	// in which case the expected results must all match across all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only operate over public document(s).
	//
	// If an Identity is provided and the collection has a policy, then can
	// operate over private document(s) that are owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[state.Identity]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// Materialized views are automatically refreshed immediately before executing this Request, unless
	// this property is set to true.
	DoNotRefreshViews bool

	// OperationName sets the operation name option for the request.
	OperationName immutable.Option[string]

	// Variables sets the variables option for the request.
	Variables immutable.Option[map[string]any]

	// The request to execute.
	Request string

	// The expected (data) results of the issued request.
	Results map[string]any

	// NonOrderedResults specifies that the results set doesn't need to care about the ordering of the items.
	NonOrderedResults bool

	// Asserter is an optional custom result asserter.
	Asserter ResultAsserter

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*Request)(nil)
var _ Stateful = (*Request)(nil)

// Execute executes the request action.
func (a *Request) Execute() {
	var expectedErrorRaised bool
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)

nodeLoop:
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		// Check if a transaction is attached to this action. If so, we will be using it.
		hadTxn := a.TransactionID.HasValue()
		var txn client.Txn
		var err error
		if hadTxn {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
		}

		reqOption := options.ExecRequest()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			reqOption.SetIdentity(identOption.Value())
		}
		if a.OperationName.HasValue() {
			reqOption.SetOperationName(a.OperationName.Value())
		}
		if a.Variables.HasValue() {
			reqOption.SetVariables(resolveVariables(a.s, a.Variables.Value()))
		}

		if !a.DoNotRefreshViews && !expectedErrorRaised {
			expectedErrorRaised = refreshViews(a.s, node, identOption, a.ExpectedError)
			if expectedErrorRaised {
				continue nodeLoop
			}
		}

		request := replace(a.s, nodeID, a.Request)
		// If we have a transaction, we will use it here. Otherwise we use the node.
		var result *client.RequestResult
		if hadTxn {
			result = txn.ExecRequest(a.s.Ctx, request, reqOption)
		} else {
			result = node.ExecRequest(a.s.Ctx, request, reqOption)
		}

		expectedErrorRaised = assertRequestResults(
			a.s,
			&result.GQL,
			a.Results,
			a.ExpectedError,
			a.Asserter,
			nodeID,
			!a.NonOrderedResults,
		)
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}

// resolveVariables resolves any DocIndex values in the variables map to their
// corresponding document ID strings.
func resolveVariables(s *state.State, vars map[string]any) map[string]any {
	resolved := make(map[string]any, len(vars))
	for k, v := range vars {
		if index, ok := v.(DocIndex); ok {
			s.DocIDsLock.RLock()
			docID := s.DocIDs[index.CollectionIndex][index.Index]
			s.DocIDsLock.RUnlock()
			resolved[k] = docID.String()
		} else {
			resolved[k] = v
		}
	}
	return resolved
}
