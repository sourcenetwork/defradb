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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// NewIndex will attempt to make a new secondary index for the given collection
// using the collection api.
type NewIndex struct {
	stateful

	// NodeID may hold the ID (index) of a node to make the new secondary index on.
	//
	// If a value is not provided the index will be made on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which this index should be added.
	CollectionID int

	// The name of the index to add. If not provided, one will be generated.
	IndexName string

	// The name of the field to index. Used only for single field indexes.
	// It's a convenience field so that tests don't have to add a slice
	// of [IndexedField] when only a single field index is needed.
	FieldName string

	// The fields to index. Used only for composite indexes.
	Fields []client.IndexedFieldDescription

	// If Unique is true, the index will be added as a unique index.
	//
	// Deprecated: use Ordered.
	Unique bool

	// Ordered, when set, carries the ordered index config.
	Ordered *client.OrderedIndexDescription

	// Vector, when set, creates a vector (ANN) index instead of a secondary index. It carries the
	// algorithm, metric, dimensions and params. Used to create a vector index through the index API on
	// a collection that already holds documents, so the async backfill builds the graph.
	Vector *client.VectorIndexDescription

	// Async, when true, returns without waiting for the backfill. The index stays building until
	// the worker completes it.
	//
	// By default the action waits for the build to reach ready or failed, so a following query
	// sees a built index. Set Async true to observe the building window, e.g. with a following
	// ListIndexes asserting an in-progress status.
	Async bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	//
	// NewIndex is async: a backfill failure is not returned here, it surfaces as a failed index
	// status. Use ListIndexes with ExpectedStatuses to assert a build failure; ExpectedError
	// only covers synchronous validation errors (e.g. a duplicate index name).
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*NewIndex)(nil)
var _ Stateful = (*NewIndex)(nil)

func (a *NewIndex) Execute() {
	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		node := a.s.Nodes[nodeID]

		// Check if a transaction is attached to this action. If so, we will be using it.
		txnOption := immutable.None[client.Txn]()
		if a.TransactionID.HasValue() {
			txn, err := a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
		}

		collections, err := GetCollectionsCanonically(a.s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}

		collection := collections[a.CollectionID]

		indexDesc := client.NewIndexRequest{
			Name: a.IndexName,
		}

		if a.FieldName != "" {
			//nolint:staticcheck // the action exposes both spellings so tests can cover each
			indexDesc.Fields = []client.IndexedFieldDescription{
				{
					Name: a.FieldName,
				},
			}
		} else if len(a.Fields) > 0 {
			for i := range a.Fields {
				indexDesc.Fields = append(indexDesc.Fields, client.IndexedFieldDescription{
					Name:       a.Fields[i].Name,
					Descending: a.Fields[i].Descending,
				})
			}
		}

		//nolint:staticcheck // the action exposes both spellings so tests can cover each
		indexDesc.Unique = a.Unique
		indexDesc.Ordered = a.Ordered
		indexDesc.Vector = a.Vector

		opts := options.NewCollectionIndex()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		desc, err := collection.NewIndex(a.s.Ctx, indexDesc, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		// Unless the test wants to observe the building window, wait for the build so a following
		// query sees a built index. With an explicit transaction the record is not committed until
		// the caller commits, so there is nothing to wait for yet; the test waits after its commit.
		if err == nil && !a.Async && !a.TransactionID.HasValue() {
			waitForIndexBuilt(a.s, collection, desc.ID, listIndexesOptions(a.s, node))
		}
	}
}

// indexBuildTimeout bounds the wait for a background build or drop to finish. The poll returns as
// soon as the state clears, so this only elapses on a real hang; it is sized for a slow backfill of
// a large collection in CI.
const indexBuildTimeout = 10 * time.Second

// waitForIndexBuilt blocks until the given index leaves the building state (ready or failed). It
// polls ListIndexes rather than the action event bus so it composes with explicit Wait actions that
// consume the shared subscription.
func waitForIndexBuilt(
	s *state.State,
	collection client.Collection,
	indexID uint32,
	opts options.Enumerable[options.ListCollectionIndexesOptions],
) {
	require.Eventually(s.T, func() bool {
		results, err := collection.ListIndexes(s.Ctx, opts)
		if err != nil {
			return false
		}
		for _, r := range results {
			if r.Description.ID != indexID {
				continue
			}
			// building → BackfillIndexAction + InProgress; anything else is terminal.
			return r.Execution.Action != client.BackfillIndexAction ||
				r.Execution.Status != client.InProgressActionStatus
		}
		// The index is gone from the listing only if it was dropped; treat as settled.
		return true
	}, indexBuildTimeout, time.Millisecond, "timed out waiting for index %d to finish building", indexID)
}
