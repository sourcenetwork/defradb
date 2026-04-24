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
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/state"
)

type UpdateDoc struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only update public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also update private document(s) that are owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection in which this document exists.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int

	// The document update, in JSON string format. Will only update the properties
	// provided.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Skip waiting for an update event on the local event bus.
	//
	// This should only be used for tests that do not correctly
	// publish an update event to the local event bus.
	SkipLocalUpdateEvent bool

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*UpdateDoc)(nil)
var _ Stateful = (*UpdateDoc)(nil)

func (a *UpdateDoc) Execute() {
	var mutation func(
		*state.State,
		*UpdateDoc,
		client.TxnStore,
		int,
		client.Collection,
		immutable.Option[client.Txn],
	) error

	switch state.ActiveMutationType {
	case state.CollectionSaveMutationType:
		mutation = updateDocViaColSave
	case state.CollectionNamedMutationType:
		mutation = updateDocViaColUpdate
	case state.GQLRequestMutationType:
		mutation = updateDocViaGQL
	default:
		a.s.T.Fatalf("invalid mutationType: %v", state.ActiveMutationType)
	}

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	var expectedErrorRaised bool
	doNotWaitForUpdate := false
	var err error

	for index, node := range nodes {
		nodeID := nodeIDs[index]

		txnOption := immutable.None[client.Txn]()
		if a.TransactionID.HasValue() {
			txn, err := a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
			doNotWaitForUpdate = true // if using txn, we skip local update wait
		}

		collections := GetCanonicallyOrderedCollections(a.s, node, txnOption)
		collection := collections[a.CollectionID]

		err = withRetryOnNode(node, func() error {
			return mutation(a.s, a, node, nodeID, collection, txnOption)
		})

		expectedErrorRaised = assertError(a.s.T, err, a.ExpectedError)
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

	if a.ExpectedError == "" && !a.SkipLocalUpdateEvent && !doNotWaitForUpdate {
		waitForUpdateEvents(
			a.s,
			a.NodeID,
			a.CollectionID,
			getEventsForUpdateDoc(a.s, a),
			immutable.None[state.Identity](),
		)
	}
}

func updateDocViaColSave(
	s *state.State,
	action *UpdateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) error {
	ctx := s.Ctx
	if txn.HasValue() {
		ctx = db.InitContext(s.Ctx, txn.Value())
	}

	s.DocIDsLock.RLock()
	docID := s.DocIDs[action.CollectionID][action.DocID]
	s.DocIDsLock.RUnlock()

	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	getOpts := options.GetDocument()
	if identOption.HasValue() {
		getOpts.SetIdentity(identOption.Value())
	}
	doc, err := collection.GetDocument(ctx, docID, getOpts.SetShowDeleted(true))
	if err != nil {
		return err
	}
	err = doc.SetWithJSON(ctx, []byte(action.Doc))
	if err != nil {
		return err
	}

	saveOpts := options.SaveDocument()
	if identOption.HasValue() {
		saveOpts.SetIdentity(identOption.Value())
	}
	return collection.SaveDocument(ctx, doc, saveOpts)
}

func updateDocViaColUpdate(
	s *state.State,
	action *UpdateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) error {
	ctx := s.Ctx
	if txn.HasValue() {
		ctx = db.InitContext(s.Ctx, txn.Value())
	}

	s.DocIDsLock.RLock()
	docID := s.DocIDs[action.CollectionID][action.DocID]
	s.DocIDsLock.RUnlock()

	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	getOpts := options.GetDocument()
	if identOption.HasValue() {
		getOpts.SetIdentity(identOption.Value())
	}
	doc, err := collection.GetDocument(ctx, docID, getOpts.SetShowDeleted(true))
	if err != nil {
		return err
	}
	err = doc.SetWithJSON(ctx, []byte(action.Doc))
	if err != nil {
		return err
	}

	updateOpts := options.UpdateDocument()
	if identOption.HasValue() {
		updateOpts.SetIdentity(identOption.Value())
	}
	return collection.UpdateDocument(ctx, doc, updateOpts)
}

func updateDocViaGQL(
	s *state.State,
	action *UpdateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) error {
	ctx := s.Ctx
	hadTxn := txn.HasValue()
	if hadTxn {
		ctx = db.InitContext(s.Ctx, txn.Value())
	}

	s.DocIDsLock.RLock()
	docID := s.DocIDs[action.CollectionID][action.DocID]
	s.DocIDsLock.RUnlock()

	input, err := jsonToGQL(action.Doc)
	require.NoError(s.T, err)

	request := fmt.Sprintf(
		`mutation {
			update_%s(docID: "%s", input: %s) {
				_docID
			}
		}`,
		collection.Name(),
		docID.String(),
		input,
	)

	reqOption := options.ExecRequest()
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	if identOption.HasValue() {
		reqOption.SetIdentity(identOption.Value())
	}

	var result *client.RequestResult
	if hadTxn {
		result = txn.Value().ExecRequest(ctx, request, reqOption)
	} else {
		result = node.ExecRequest(ctx, request, reqOption)
	}
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}
	return nil
}

// getEventsForUpdateDoc returns a map of docIDs that should be
// published to the local event bus after an UpdateDoc action.
func getEventsForUpdateDoc(s *state.State, action *UpdateDoc) map[string]struct{} {
	s.DocIDsLock.RLock()
	docID := s.DocIDs[action.CollectionID][action.DocID]
	s.DocIDsLock.RUnlock()

	docMap := make(map[string]any)
	err := json.Unmarshal([]byte(action.Doc), &docMap)
	require.NoError(s.T, err)

	return map[string]struct{}{
		docID.String(): {},
	}
}
