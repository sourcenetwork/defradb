// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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

type CreateDoc struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this create to.
	//
	// If a value is not provided the document will be created in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided the created document(s) will be public.
	//
	// If an Identity is provided and the collection has a policy, then the
	// created document(s) will be owned by this Identity.
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
}

var _ Action = (*CreateDoc)(nil)
var _ Stateful = (*CreateDoc)(nil)

func (a *CreateDoc) Execute() {
	if a.DocMap != nil {
		substituteRelations(a.s, a)
	}

	var mutation func(*CreateDoc, client.TxnStore, int, client.Collection) ([]client.DocID, error)
	switch state.ActiveMutationType {
	case state.CollectionSaveMutationType:
		mutation = createDocViaColSave
	case state.CollectionNamedMutationType:
		mutation = createDocViaColCreate
	case state.GQLRequestMutationType:
		mutation = createDocViaGQL
	default:
		a.s.T.Fatalf("invalid mutationType: %v", state.ActiveMutationType)
	}

	var expectedErrorRaised bool
	var docIDs []client.DocID

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := a.s.Nodes[nodeID].Collections[a.CollectionID]
		err := withRetryOnNode(
			node,
			func() error {
				var err error
				docIDs, err = mutation(
					a,
					node,
					nodeID,
					collection,
				)
				return err
			},
		)
		expectedErrorRaised = assertError(a.s.T, err, a.ExpectedError)
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

	if a.ExpectedError == "" && !a.DoNotWaitForEvent {
		waitForUpdateEvents(a.s, a.NodeID, a.CollectionID, docIDMap, a.Identity)
	}
}

func createDocViaColSave(
	a *CreateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
	txn, err := a.s.GetTransaction(node, immutable.None[int]())
	if err != nil {
		return nil, err
	}

	ctx := db.InitContext(a.s.Ctx, txn)

	docs, err := parseCreateDocs(ctx, a, collection)
	if err != nil {
		return nil, err
	}
	docIDs := make([]client.DocID, len(docs))
	for i, doc := range docs {
		err := collection.Save(ctx, doc, makeDocSaveOptions(a.s, a, nodeIndex)...)
		if err != nil {
			return nil, err
		}
		docIDs[i] = doc.ID()
	}
	return docIDs, nil
}

func createDocViaColCreate(
	a *CreateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
	txn, err := a.s.GetTransaction(node, immutable.None[int]())
	if err != nil {
		return nil, err
	}

	ctx := db.InitContext(a.s.Ctx, txn)

	docs, err := parseCreateDocs(ctx, a, collection)
	if err != nil {
		return nil, err
	}

	switch {
	case len(docs) > 1:
		err := collection.CreateMany(ctx, docs, makeDocCreateOptions(a.s, a, nodeIndex)...)
		if err != nil {
			return nil, err
		}

	default:
		err := collection.Create(ctx, docs[0], makeDocCreateOptions(a.s, a, nodeIndex)...)
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

func createDocViaGQL(
	a *CreateDoc,
	node client.TxnStore,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
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

	key := fmt.Sprintf("create_%s", collection.Name())
	req := fmt.Sprintf(`mutation { %s(%s) { _docID } }`, key, params)

	txn, err := a.s.GetTransaction(node, immutable.None[int]())
	if err != nil {
		return nil, err
	}

	ctx := db.InitContext(a.s.Ctx, txn)

	reqOption := options.ExecRequest()
	identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeIndex)
	if identOption.HasValue() {
		reqOption.SetIdentity(identOption.Value())
	}

	result := node.ExecRequest(ctx, req, reqOption)
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
	action *CreateDoc,
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

// parseCreateDocs parses and returns documents from a CreateDoc action.
func parseCreateDocs(ctx context.Context, action *CreateDoc, collection client.Collection) ([]*client.Document, error) {
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
	action *CreateDoc,
	nodeIndex int,
) []options.Enumerable[options.CollectionSaveOptions] {
	opts := options.CollectionSave().
		SetEncryptDoc(action.IsDocEncrypted).
		SetEncryptedFields(action.EncryptedFields)
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	return []options.Enumerable[options.CollectionSaveOptions]{opts}
}

func makeDocCreateOptions(
	s *state.State,
	action *CreateDoc,
	nodeIndex int,
) []options.Enumerable[options.CollectionCreateOptions] {
	opts := options.CollectionCreate().
		SetEncryptDoc(action.IsDocEncrypted).
		SetEncryptedFields(action.EncryptedFields)
	identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeIndex)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	return []options.Enumerable[options.CollectionCreateOptions]{opts}
}
