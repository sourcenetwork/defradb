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
	"fmt"
	"testing"
	"time"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
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
	coordinator      *ReplicationCoordinator

	// Test data
	docID        string
	collectionID string
	fieldName    string
	peerID       string
	encKey       []byte
}

// newTestSetup creates a new test setup with all mocks initialized
func newTestSetup(t *testing.T) *testSetup {
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

		// Default test data
		docID:        "bae-63c10140-a59a-5a7f-85d1-269e2c3841a6",
		collectionID: "test-collection",
		fieldName:    "age",
		peerID:       "peer1",
		encKey:       []byte("test-encryption-key-32-bytes-!"),
	}

	// Setup default mock expectations for coordinator creation
	setup.mockDB.EXPECT().Events().Return(setup.mockEventBus)
	setup.mockDB.EXPECT().MaxTxnRetries().Return(3)

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
	// Mock collection with encrypted index
	mockCollection := &mockSimpleCollection{
		name:         "TestCollection",
		collectionID: s.collectionID,
		encryptedIndexes: []client.EncryptedIndexDescription{
			{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
		},
	}

	// Mock GetCollections call
	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	// Mock P2P replicator IDs
	s.mockP2P.EXPECT().GetReplicatorsIDs(s.collectionID).Return([]string{s.peerID})

	// Expect SE artifact push to peer
	s.mockStorageProto.EXPECT().SendRequest(
		mock.Anything,
		mock.MatchedBy(func(req PushSEArtifactsRequest) bool {
			return req.CollectionID == s.collectionID && len(req.Artifacts) > 0
		}),
		s.peerID,
		false, // isRetry = false
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
	time.Sleep(10 * time.Millisecond)
	require.Empty(s.t, s.mockStorageProto.Calls, "No SE artifacts should be pushed")
}

// expectCollectionFound sets up basic collection mock expectation
func (s *testSetup) expectCollectionFound() {
	mockCollection := s.createMockCollection()
	s.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)
}

// createMockCollection creates a configurable mock collection
func (s *testSetup) createMockCollection() *mockSimpleCollection {
	return &mockSimpleCollection{
		name:         "TestCollection",
		collectionID: s.collectionID,
		encryptedIndexes: []client.EncryptedIndexDescription{
			{FieldName: s.fieldName, Type: client.EncryptedIndexTypeEquality},
		},
	}
}

// createNonCompositeBlock creates a non-composite block for testing
func (s *testSetup) createNonCompositeBlock() []byte {
	// Create a field-level block (non-composite)
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

// Close cleans up the coordinator and verifies all expectations
func (s *testSetup) close() {
	if s.coordinator != nil {
		s.coordinator.Close()
	}
	s.mockDB.AssertExpectations(s.t)
	s.mockP2P.AssertExpectations(s.t)
	s.mockStorageProto.AssertExpectations(s.t)
}

func (setup *testSetup) makeUpdateEvent() event.Update {
	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        setup.createValidCompositeBlock(),
	}
	return updateEvent
}

func TestReplicationCoordinator_WhenUpdateEventReceived_ShouldPushSEArtifactsToPeers(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.expectSEArtifactPush()

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForArtifactPush(func(req *PushSEArtifactsRequest) bool {
		return req.CollectionID == setup.collectionID && len(req.Artifacts) > 0
	})
}

func TestReplicationCoordinator_WhenBlockFailsToDeserialize_ShouldReturnError(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        []byte("invalid-block-data"),
	}
	setup.publishEvent(event.UpdateName, updateEvent)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNonCompositeBlock_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	updateEvent := event.Update{
		DocID:        setup.docID,
		CollectionID: setup.collectionID,
		Block:        setup.createNonCompositeBlock(),
	}
	setup.publishEvent(event.UpdateName, updateEvent)

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenGetCollectionsFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database error"))

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenCollectionNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{}, nil)

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenInvalidDocID_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	setup.docID = "invalid-doc-id" // Set invalid docID
	setup.expectCollectionFound()

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenDocumentNotFound_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.docNotFound = true
	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenDocumentGetFails_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.getError = fmt.Errorf("database error")
	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

func TestReplicationCoordinator_WhenNoEncryptedIndexes_ShouldNotPushToPeer(t *testing.T) {
	setup := newTestSetup(t)
	defer setup.close()

	mockCollection := setup.createMockCollection()
	mockCollection.encryptedIndexes = []client.EncryptedIndexDescription{} // Empty list
	setup.mockDB.EXPECT().GetCollections(mock.Anything, mock.Anything).Return([]client.Collection{mockCollection}, nil)

	setup.publishEvent(event.UpdateName, setup.makeUpdateEvent())

	setup.waitForNoCalls()
}

// newMockProto creates a mock proto for testing
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

func (m *mockProto[Req, Rep]) SendRequest(ctx context.Context, req Req, peerID string, isRetry bool) (Rep, error) {
	args := m.Called(ctx, req, peerID, isRetry)
	return args.Get(0).(Rep), args.Error(1)
}

func (m *mockProto[Req, Rep]) EXPECT() *mockProtoExpectation[Req, Rep] {
	return &mockProtoExpectation[Req, Rep]{mock: &m.Mock}
}

type mockProtoExpectation[Req, Rep any] struct {
	mock *mock.Mock
}

func (e *mockProtoExpectation[Req, Rep]) SendRequest(ctx, req, peerID, isRetry interface{}) *mock.Call {
	return e.mock.On("SendRequest", ctx, req, peerID, isRetry)
}

// mockEventBus implements a simple event bus for testing
type mockEventBus struct {
	messages chan event.Message
	subs     map[event.Subscription]chan event.Message
}

func (m *mockEventBus) Publish(msg event.Message) {
	// Send to all subscribers
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

// createValidCompositeBlock creates a proper CBOR-encoded composite block using the pattern from block_test.go
func createValidCompositeBlock(t *testing.T, docID, collectionID, fieldName string) []byte {
	// Create linking system for storing blocks (copied from block_test.go)
	store := &memstore.Store{}
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(store)
	lsys.SetWriteStorage(store)

	// Generate field block first (copied from block_test.go pattern)
	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{
			LWWDelta: &crdt.LWWDelta{
				DocID:           []byte(docID),
				FieldName:       fieldName,
				Priority:        1,
				SchemaVersionID: collectionID,
				Data:            []byte("21"), // Test field value
			},
		},
	}
	fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), fieldBlock.GenerateNode())
	require.NoError(t, err)

	// Create composite block (copied from block_test.go pattern)
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

	// Marshal to get proper CBOR bytes
	blockBytes, err := compositeBlock.Marshal()
	require.NoError(t, err)
	return blockBytes
}

// mockSimpleCollection is a configurable collection mock for testing
type mockSimpleCollection struct {
	name             string
	collectionID     string
	encryptedIndexes []client.EncryptedIndexDescription

	// Error simulation fields
	docNotFound bool
	getError    error
}

func (m *mockSimpleCollection) Name() string         { return m.name }
func (m *mockSimpleCollection) VersionID() string    { return "v1" }
func (m *mockSimpleCollection) CollectionID() string { return m.collectionID }
func (m *mockSimpleCollection) Version() client.CollectionVersion {
	return client.CollectionVersion{
		Name:         m.name,
		CollectionID: m.collectionID,
		Fields: []client.CollectionFieldDescription{
			{
				Name: "age",
				Kind: client.FieldKind_NILLABLE_INT,
			},
		},
		EncryptedIndexes: m.encryptedIndexes,
	}
}
func (m *mockSimpleCollection) GetEncryptedIndexes(context.Context) ([]client.EncryptedIndexDescription, error) {
	return m.encryptedIndexes, nil
}

// Stub out all other Collection methods for this simple mock
func (m *mockSimpleCollection) Create(context.Context, *client.Document, ...client.DocCreateOption) error {
	return nil
}
func (m *mockSimpleCollection) CreateMany(context.Context, []*client.Document, ...client.DocCreateOption) error {
	return nil
}
func (m *mockSimpleCollection) Update(context.Context, *client.Document) error { return nil }
func (m *mockSimpleCollection) Save(context.Context, *client.Document, ...client.DocCreateOption) error {
	return nil
}
func (m *mockSimpleCollection) Delete(context.Context, client.DocID) (bool, error) { return false, nil }
func (m *mockSimpleCollection) Exists(context.Context, client.DocID) (bool, error) { return false, nil }
func (m *mockSimpleCollection) UpdateWithFilter(context.Context, any, string) (*client.UpdateResult, error) {
	return nil, nil
}
func (m *mockSimpleCollection) DeleteWithFilter(context.Context, any) (*client.DeleteResult, error) {
	return nil, nil
}
func (m *mockSimpleCollection) Get(context.Context, client.DocID, bool) (*client.Document, error) {
	// Simulate errors if configured
	if m.docNotFound {
		return nil, client.ErrDocumentNotFoundOrNotAuthorized
	}
	if m.getError != nil {
		return nil, m.getError
	}

	// Return a simple document for testing
	doc, err := client.NewDocFromMap(map[string]any{"age": 21}, m.Version())
	return doc, err
}
func (m *mockSimpleCollection) GetAllDocIDs(context.Context) (<-chan client.DocIDResult, error) {
	return nil, nil
}
func (m *mockSimpleCollection) CreateIndex(context.Context, client.IndexCreateRequest) (client.IndexDescription, error) {
	return client.IndexDescription{}, nil
}
func (m *mockSimpleCollection) DropIndex(context.Context, string) error { return nil }
func (m *mockSimpleCollection) GetIndexes(context.Context) ([]client.IndexDescription, error) {
	return nil, nil
}
func (m *mockSimpleCollection) CreateEncryptedIndex(context.Context, client.EncryptedIndexCreateRequest) (client.EncryptedIndexDescription, error) {
	return client.EncryptedIndexDescription{}, nil
}
func (m *mockSimpleCollection) DeleteEncryptedIndex(context.Context, string) error { return nil }
