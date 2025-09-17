// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"testing"
	"time"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"

	"github.com/sourcenetwork/defradb/client"
	clientmocks "github.com/sourcenetwork/defradb/client/mocks"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/se/mocks"
)

// testSetup holds all test mocks and utilities for ReplicationCoordinator testing
type testSetup struct {
	t                *testing.T
	mockDB           *mocks.DB
	mockP2P          *mocks.P2P
	mockStorageProto *mockProto[PushSEArtifactsRequest, PushSEArtifactsReply]
	mockQueryProto   *mockProto[QuerySEArtifactsRequest, QuerySEArtifactsReply]
	mockEventBus     *mockEventBus
	coordinator      *Coordinator
	rootstore        *memory.Datastore

	// Test data
	docID        string
	collectionID string
	fieldName    string
	peerID       string
	encKey       []byte
}

// newTestSetup creates a new test setup with all mocks initialized
func newTestSetup(t *testing.T) *testSetup {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	setup := &testSetup{
		t:                t,
		mockDB:           mocks.NewDB(t),
		mockP2P:          mocks.NewP2P(t),
		mockStorageProto: newMockProto[PushSEArtifactsRequest, PushSEArtifactsReply](t),
		mockQueryProto:   newMockProto[QuerySEArtifactsRequest, QuerySEArtifactsReply](t),
		mockEventBus: &mockEventBus{
			messages: make(chan event.Message, 10),
			subs:     make(map[event.Subscription]chan event.Message),
		},
		rootstore: rootstore,

		docID:        "bae-63c10140-a59a-5a7f-85d1-269e2c3841a6",
		collectionID: "test-collection",
		fieldName:    "age",
		peerID:       "peer1",
		encKey:       []byte("test-encryption-key-32-bytes-!"),
	}

	setup.mockDB.EXPECT().Events().Return(setup.mockEventBus).Maybe()
	setup.mockDB.EXPECT().MaxTxnRetries().Return(3).Maybe()
	setup.mockDB.EXPECT().Rootstore().Return(rootstore).Maybe()

	setup.createCoordinator()

	return setup
}

// createCoordinator creates the ReplicationCoordinator with all mocks
func (s *testSetup) createCoordinator() {
	rc, err := newReplicationCoordinator(
		s.mockDB,
		s.mockP2P,
		s.encKey,
		s.mockStorageProto,
		s.mockQueryProto,
	)
	require.NoError(s.t, err)
	s.coordinator = rc
}

// expectSEArtifactPush sets up expectation for SE artifact push to peer
func (s *testSetup) expectSEArtifactPush() {
	mockCollection := s.createMockCollectionWithDocument()

	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	s.mockGetReplicatorsIDs([]string{s.peerID})

	// Expect SE artifact push to peer
	s.mockStorageProto.EXPECT().SendRequest(
		mock.Anything,
		mock.MatchedBy(func(req PushSEArtifactsRequest) bool {
			return req.CollectionID == s.collectionID && len(req.Artifacts) > 0
		}),
		s.peerID,
	).Return(PushSEArtifactsReply{}, nil)
}

// publishEvent publishes any event with the given name and data to the event bus
func (s *testSetup) publishEvent(name event.Name, evt any) {
	s.mockEventBus.Publish(event.NewMessage(name, evt))
}

// createValidCompositeBlock creates a proper CBOR-encoded composite block
func (s *testSetup) createValidCompositeBlock() []byte {
	return createValidCompositeBlock(s.t, s.docID, s.collectionID, s.fieldName)
}

// waitForArtifactPush waits for SE artifacts to be pushed and validates with custom assertion
func (s *testSetup) waitForArtifactPush(validate func(*PushSEArtifactsRequest) bool) {
	require.Eventually(s.t, func() bool {
		calls := s.mockStorageProto.Calls
		for _, call := range calls {
			if call.Method == "SendRequest" && len(call.Arguments) > 1 {
				if req, ok := call.Arguments[1].(PushSEArtifactsRequest); ok {
					return validate(&req)
				}
			}
		}
		return false
	}, time.Second, 10*time.Millisecond, "SE artifacts should be pushed to replicator with expected data")
}

// waitForNoCalls verifies that no calls were made to the storage protocol
func (s *testSetup) waitForNoCalls() {
	// Wait a bit to ensure no async calls happen
	time.Sleep(20 * time.Millisecond)
	require.Empty(s.t, s.mockStorageProto.Calls, "No SE artifacts should be pushed")
}

func (s *testSetup) mockGetReplicatorsIDs(peers []string) {
	s.mockP2P.EXPECT().GetReplicatorsIDs(s.collectionID).Return(peers).Maybe()
}

func (s *testSetup) mockGetCollections(m ...*clientmocks.Collection) {
	cols := make([]client.Collection, len(m))
	for i, col := range m {
		cols[i] = col
	}
	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Maybe().Return(cols, nil)
}

func getCollectionFieldsDescriptions() []client.CollectionFieldDescription {
	return []client.CollectionFieldDescription{
		{
			Name: "age",
			Kind: client.FieldKind_NILLABLE_INT,
		},
	}
}

func (s *testSetup) createEncryptedIndexesDescriptions() []client.EncryptedIndexDescription {
	return []client.EncryptedIndexDescription{
		{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
	}
}

func (s *testSetup) createCollectionVersion() client.CollectionVersion {
	return client.CollectionVersion{
		Name:             "TestCollection",
		CollectionID:     s.collectionID,
		Fields:           getCollectionFieldsDescriptions(),
		EncryptedIndexes: s.createEncryptedIndexesDescriptions(),
	}
}

// createMockCollection creates a configurable mock collection
func (s *testSetup) createMockCollection() *clientmocks.Collection {
	mockCollection := clientmocks.NewCollection(s.t)

	mockCollection.EXPECT().Name().Return("TestCollection").Maybe()
	mockCollection.EXPECT().CollectionID().Return(s.collectionID).Maybe()
	mockCollection.EXPECT().VersionID().Return("v1").Maybe()

	mockCollection.EXPECT().GetEncryptedIndexes(mock.Anything).Return(
		[]client.EncryptedIndexDescription{
			{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
		}, nil).Maybe()

	mockCollection.EXPECT().Version().Return(s.createCollectionVersion()).Maybe()

	return mockCollection
}

// createMockCollectionWithDocument creates a mock collection that returns a successful Get
func (s *testSetup) createMockCollectionWithDocument() *clientmocks.Collection {
	mockCollection := s.createMockCollection()

	// Setup Get method with default return
	mockCollection.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, docID client.DocID, showDeleted bool) (*client.Document, error) {
			doc, err := client.NewDocFromMap(map[string]any{"age": 21}, mockCollection.Version())
			return doc, err
		}).Maybe()

	return mockCollection
}

// createNonCompositeBlock creates a non-composite block for testing
func (s *testSetup) createNonCompositeBlock() []byte {
	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:           []byte(s.docID),
				FieldName:       s.fieldName,
				Priority:        1,
				SchemaVersionID: s.collectionID,
				Data:            []byte("21"),
			},
		},
	}

	blockBytes, err := fieldBlock.Marshal()
	require.NoError(s.t, err)
	return blockBytes
}

// close cleans up the coordinator and verifies all expectations
func (s *testSetup) close() {
	if s.coordinator != nil {
		s.coordinator.Close()
	}
	s.mockDB.AssertExpectations(s.t)
	s.mockP2P.AssertExpectations(s.t)
	s.mockStorageProto.AssertExpectations(s.t)
}

func (s *testSetup) makeUpdateEvent() event.Update {
	updateEvent := event.Update{
		DocID:        s.docID,
		CollectionID: s.collectionID,
		Block:        s.createValidCompositeBlock(),
	}
	return updateEvent
}

// createValidCompositeBlock creates a proper CBOR-encoded composite block using the pattern from block_test.go
func createValidCompositeBlock(t *testing.T, docID, collectionID, fieldName string) []byte {
	store := &memstore.Store{}
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(store)
	lsys.SetWriteStorage(store)

	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:           []byte(docID),
				FieldName:       fieldName,
				Priority:        1,
				SchemaVersionID: collectionID,
				Data:            []byte("21"),
			},
		},
	}
	fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), fieldBlock.GenerateNode())
	require.NoError(t, err)

	compositeBlock := coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{
				DocID:           []byte(docID),
				Priority:        1,
				SchemaVersionID: collectionID,
				Status:          1,
			},
		},
		Links: []coreblock.DAGLink{
			{
				Name: fieldName,
				Link: fieldBlockLink.(cidlink.Link),
			},
		},
	}

	blockBytes, err := compositeBlock.Marshal()
	require.NoError(t, err)
	return blockBytes
}

func newMockProto[Req, Rep any](t *testing.T) *mockProto[Req, Rep] {
	return &mockProto[Req, Rep]{
		Mock: mock.Mock{},
		t:    t,
	}
}

type mockProto[Req, Rep any] struct {
	mock.Mock
	t testing.TB
}

func (m *mockProto[Req, Rep]) SendRequest(ctx context.Context, req Req, peerID string) (Rep, error) {
	args := m.Called(ctx, req, peerID)
	return args.Get(0).(Rep), args.Error(1)
}

func (m *mockProto[Req, Rep]) EXPECT() *mockProtoExpectation[Req, Rep] {
	return &mockProtoExpectation[Req, Rep]{mock: &m.Mock}
}

type mockProtoExpectation[Req, Rep any] struct {
	mock *mock.Mock
}

func (e *mockProtoExpectation[Req, Rep]) SendRequest(ctx, req, peerID any) *mock.Call {
	return e.mock.On("SendRequest", ctx, req, peerID)
}

type mockEventBus struct {
	messages chan event.Message
	subs     map[event.Subscription]chan event.Message
}

func (m *mockEventBus) Publish(msg event.Message) {
	for _, ch := range m.subs {
		select {
		case ch <- msg:
		default:
			// Don't block if channel is full
		}
	}
}

func (m *mockEventBus) Subscribe(events ...event.Name) (event.Subscription, error) {
	ch := make(chan event.Message, 10)
	sub := &mockSubscription{ch: ch}
	m.subs[sub] = ch
	return sub, nil
}

func (m *mockEventBus) Unsubscribe(sub event.Subscription) {
	if ch, exists := m.subs[sub]; exists {
		close(ch)
		delete(m.subs, sub)
	}
}

func (m *mockEventBus) Close() {
	for _, ch := range m.subs {
		close(ch)
	}
}

// mockSubscription implements event.Subscription for testing
type mockSubscription struct {
	ch chan event.Message
}

func (m *mockSubscription) Message() <-chan event.Message {
	return m.ch
}
