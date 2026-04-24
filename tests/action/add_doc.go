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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

type AddDoc struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this add to.
	//
	// If a value is not provided the document will be added in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided the added document(s) will be public.
	//
	// If an Identity is provided and the collection has a policy, then the
	// added document(s) will be owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// Specifies whether the document should be encrypted.
	IsDocEncrypted bool

	// Individual fields of the document to encrypt.
	EncryptedFields []string

	// The collection in which this document should be created.
	CollectionID int

	// The document to create, in JSON string format.
	//
	// If [DocMap] is provided this value will be ignored.
	Doc string

	// The document to create, in map format.
	//
	// If this is provided [Doc] will be ignored.
	DocMap map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// If this property true, then the action will not wait for the event(s) that it triggers
	// to be broadcasted.
	//
	// This was introduced as the function used to wait for events currently assumes that a single
	// action will be executed at any given moment.  This is no longer true for all tests.
	//
	// Setting this property to true whilst testing P2P functionality will probably result in a
	// flaky test.
	DoNotWaitForEvent bool

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]

	// If the given error is received, ignore the error and pretend the action succeeded.
	IgnoreError string
}

var _ Action = (*AddDoc)(nil)
var _ Stateful = (*AddDoc)(nil)

func (a *AddDoc) Execute() {
	hadTxn := a.TransactionID.HasValue()

	if a.DocMap != nil {
		substituteRelations(a.s, a)
	}

	var mutation func(
		*AddDoc,
		client.TxnStore,
		int,
		client.Collection,
		immutable.Option[client.Txn],
	) ([]client.DocID, error)
	switch state.ActiveMutationType {
	case state.CollectionSaveMutationType:
		mutation = addDocViaColSave
	case state.CollectionNamedMutationType:
		mutation = addDocViaColAdd
	case state.GQLRequestMutationType:
		mutation = addDocViaGQL
	default:
		a.s.T.Fatalf("invalid mutationType: %v", state.ActiveMutationType)
	}

	var expectedErrorRaised bool
	var docIDs []client.DocID
	var collections []client.Collection

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)

	// Check if a transaction is attached to this action. If so, we will be using it.
	var txn client.Txn
	txnOption := immutable.None[client.Txn]()
	if hadTxn {
		var err error
		a.DoNotWaitForEvent = true
		txn, err = a.s.GetTransaction(a.s.Nodes[a.NodeID.Value()], a.TransactionID)
		require.NoError(a.s.T, err)
		txnOption = immutable.Some(txn)
	}

	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collections = GetCanonicallyOrderedCollections(a.s, node, txnOption)
		collection := collections[a.CollectionID]

		err := withRetryOnNode(
			node,
			func() error {
				var err error
				docIDs, err = mutation(
					a,
					node,
					nodeID,
					collection,
					txnOption,
				)
				return err
			},
		)
		if err == nil || !(len(a.IgnoreError) > 0 && strings.Contains(err.Error(), a.IgnoreError)) {
			expectedErrorRaised = assertError(a.s.T, err, a.ExpectedError)
		}
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

	a.s.DocIDsLock.Lock()
	if a.CollectionID >= len(a.s.DocIDs) {
		// Expand the slice if required, so that the document can be accessed by collection index
		a.s.DocIDs = append(a.s.DocIDs, make([][]client.DocID, a.CollectionID-len(a.s.DocIDs)+1)...)
	}
	a.s.DocIDs[a.CollectionID] = append(a.s.DocIDs[a.CollectionID], docIDs...)
	a.s.DocIDsLock.Unlock()

	docIDMap := make(map[string]struct{})
	for _, docID := range docIDs {
		docIDMap[docID.String()] = struct{}{}
	}

	// If there was an explicit transaction, then we will not be waiting for update events.
	if a.ExpectedError == "" && !a.DoNotWaitForEvent && !hadTxn {
		waitForUpdateEvents(a.s, a.NodeID, a.CollectionID, docIDMap, a.Identity)
	}
}

func addDocViaColSave(
	a *AddDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) ([]client.DocID, error) {
	ctx := a.s.Ctx
	if txn.HasValue() {
		ctx = db.InitContext(a.s.Ctx, txn.Value())
	}

	docs, err := parseAddDocs(ctx, a, collection)
	if err != nil {
		return nil, err
	}

	docIDs := make([]client.DocID, len(docs))
	for i, doc := range docs {
		err := collection.SaveDocument(ctx, doc, makeDocSaveOptions(a.s, a, nodeIndex)...)
		if err != nil {
			return nil, err
		}

		docIDs[i] = doc.ID()
	}

	return docIDs, nil
}

func addDocViaColAdd(
	a *AddDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) ([]client.DocID, error) {
	ctx := a.s.Ctx
	if txn.HasValue() {
		ctx = db.InitContext(a.s.Ctx, txn.Value())
	}

	docs, err := parseAddDocs(ctx, a, collection)
	if err != nil {
		return nil, err
	}

	switch {
	case len(docs) > 1:
		err := collection.AddManyDocuments(ctx, docs, makeDocAddOptions(a.s, a, nodeIndex)...)
		if err != nil {
			return nil, err
		}

	default:
		err := collection.AddDocument(ctx, docs[0], makeDocAddOptions(a.s, a, nodeIndex)...)
		if err != nil {
			return nil, err
		}
	}

	docIDs := make([]client.DocID, len(docs))
	for i, doc := range docs {
		docIDs[i] = doc.ID()
	}

	return docIDs, nil
}

func addDocViaGQL(
	a *AddDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
	txn immutable.Option[client.Txn],
) ([]client.DocID, error) {
	ctx := a.s.Ctx
	if txn.HasValue() {
		ctx = db.InitContext(a.s.Ctx, txn.Value())
	}

	var input string

	paramName := request.Input

	var err error
	if a.DocMap != nil {
		input, err = valueToGQL(a.DocMap)
	} else if client.IsJSONArray([]byte(a.Doc)) {
		var docMaps []map[string]any
		err = json.Unmarshal([]byte(a.Doc), &docMaps)
		require.NoError(a.s.T, err)
		input, err = arrayToGQL(docMaps)
	} else {
		input, err = jsonToGQL(a.Doc)
	}
	require.NoError(a.s.T, err)

	params := paramName + ": " + input

	if a.IsDocEncrypted {
		params = params + ", " + request.EncryptDocArgName + ": true"
	}
	if len(a.EncryptedFields) > 0 {
		params = params + ", " + request.EncryptFieldsArgName + ": [" +
			strings.Join(a.EncryptedFields, ", ") + "]"
	}

	key := fmt.Sprintf("add_%s", collection.Name())
	req := fmt.Sprintf(`mutation { %s(%s) { _docID } }`, key, params)

	reqOption := options.ExecRequest()
	identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeIndex)
	if identOption.HasValue() {
		reqOption.SetIdentity(identOption.Value())
	}

	var result *client.RequestResult
	if txn.HasValue() {
		result = txn.Value().ExecRequest(ctx, req, reqOption)
	} else {
		result = node.ExecRequest(ctx, req, reqOption)
	}
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	resultData, _ := result.GQL.Data.(map[string]any)
	resultDocs := ConvertToArrayOfMaps(a.s.T, resultData[key])

	docIDs := make([]client.DocID, len(resultDocs))
	for i, docMap := range resultDocs {
		docIDString, _ := docMap[request.DocIDFieldName].(string)
		docID, err := client.NewDocIDFromString(docIDString)
		require.NoError(a.s.T, err)
		docIDs[i] = docID
	}

	return docIDs, nil
}

// substituteRelations scans the fields defined in [action.DocMap], if any are of type [DocIndex]
// it will substitute the [DocIndex] for the corresponding document ID found in the state.
//
// If a document at that index is not found it will panic.
func substituteRelations(
	s *state.State,
	action *AddDoc,
) {
	for k, v := range action.DocMap {
		index, isIndex := v.(DocIndex)
		if !isIndex {
			continue
		}

		s.DocIDsLock.RLock()
		docID := s.DocIDs[index.CollectionIndex][index.Index]
		s.DocIDsLock.RUnlock()
		action.DocMap[k] = docID.String()
	}
}

// parseAddDocs parses and returns documents from a AddDoc action.
func parseAddDocs(ctx context.Context, action *AddDoc, collection client.Collection) ([]*client.Document, error) {
	switch {
	case action.DocMap != nil:
		val, err := client.NewDocFromMap(ctx, action.DocMap, collection.Version())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil

	case client.IsJSONArray([]byte(action.Doc)):
		return client.NewDocsFromJSON(ctx, []byte(action.Doc), collection.Version())

	default:
		val, err := client.NewDocFromJSON(ctx, []byte(action.Doc), collection.Version())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil
	}
}

func makeDocSaveOptions(
	s *state.State,
	action *AddDoc,
	nodeIndex int,
) []options.Enumerable[options.SaveDocumentOptions] {
	opts := options.SaveDocument().
		SetEncryptDoc(action.IsDocEncrypted).
		SetEncryptedFields(action.EncryptedFields)
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	return []options.Enumerable[options.SaveDocumentOptions]{opts}
}

func makeDocAddOptions(
	s *state.State,
	action *AddDoc,
	nodeIndex int,
) []options.Enumerable[options.AddDocumentOptions] {
	opts := options.AddDocument().
		SetEncryptDoc(action.IsDocEncrypted).
		SetEncryptedFields(action.EncryptedFields)
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	return []options.Enumerable[options.AddDocumentOptions]{opts}
}
