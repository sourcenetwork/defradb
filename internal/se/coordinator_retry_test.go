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

	"github.com/fxamacker/cbor/v2"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/se/mocks"
)

// retryTestSetup holds common setup for retry tests
type retryTestSetup struct {
	ctx         context.Context
	rootstore   *memory.Datastore
	coordinator *Coordinator
}

func newRetryTestSetup(t *testing.T) *retryTestSetup {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	mockDB := mocks.NewDB(t)
	mockDB.EXPECT().Rootstore().Return(rootstore).Maybe()

	return &retryTestSetup{
		ctx:       ctx,
		rootstore: rootstore,
		coordinator: &Coordinator{
			db: mockDB,
		},
	}
}

func (s *retryTestSetup) storeRetryInfo(t *testing.T, peerID string, info SERetryInfo) keys.PeerstoreSERetry {
	ps := datastore.PeerstoreFrom(s.rootstore)
	retryKey := keys.NewPeerstoreSERetry(peerID, info.CollectionID, info.DocID)
	retryData, err := cbor.Marshal(info)
	require.NoError(t, err)

	err = ps.Set(s.ctx, retryKey.Bytes(), retryData)
	require.NoError(t, err)

	return retryKey
}

func (s *retryTestSetup) getRetryInfo(t *testing.T, key keys.PeerstoreSERetry) SERetryInfo {
	ps := datastore.PeerstoreFrom(s.rootstore)
	data, err := ps.Get(s.ctx, key.Bytes())
	require.NoError(t, err)

	var info SERetryInfo
	err = cbor.Unmarshal(data, &info)
	require.NoError(t, err)

	return info
}

func createTestRetryInfo(nextRetryOffset time.Duration, numRetries int, retrying bool) SERetryInfo {
	return SERetryInfo{
		DocID:        "bae-test-doc-id",
		CollectionID: "test-collection",
		FieldNames:   []string{"field1"},
		NextRetry:    time.Now().Add(nextRetryOffset),
		NumRetries:   numRetries,
		Retrying:     retrying,
		PublicKey:    "test-public-key",
		KeyType:      "secp256k1",
	}
}

func TestProcessSERetries_WhenRetryNotDue_ShouldNotMarkAsRetrying(t *testing.T) {
	setup := newRetryTestSetup(t)

	// Test scenario: retry scheduled for future
	retryInfo := createTestRetryInfo(1*time.Hour, 0, false)
	retryKey := setup.storeRetryInfo(t, "peer-456", retryInfo)

	setup.coordinator.processSERetries(setup.ctx)

	// Verify retry was NOT modified
	updatedInfo := setup.getRetryInfo(t, retryKey)
	require.False(t, updatedInfo.Retrying, "Retry should not be marked as in progress")
	require.Equal(t, 0, updatedInfo.NumRetries, "NumRetries should not be incremented")
}

func TestProcessSERetries_WhenAlreadyRetrying_ShouldNotReprocess(t *testing.T) {
	setup := newRetryTestSetup(t)

	// Test scenario: retry is due but already in progress
	retryInfo := createTestRetryInfo(-1*time.Hour, 2, true)
	retryKey := setup.storeRetryInfo(t, "peer-789", retryInfo)

	setup.coordinator.processSERetries(setup.ctx)

	// Verify retry was NOT modified
	updatedInfo := setup.getRetryInfo(t, retryKey)
	require.True(t, updatedInfo.Retrying, "Retrying flag should remain true")
	require.Equal(t, 2, updatedInfo.NumRetries, "NumRetries should not be incremented")
}

func TestProcessSERetries_WhenNoRetries_ShouldCompleteWithoutError(t *testing.T) {
	setup := newRetryTestSetup(t)

	// Test scenario: empty peerstore (no retries to process)
	setup.coordinator.processSERetries(setup.ctx)

	// Test passes if no panic/error occurs
}

func TestProcessSERetries_WhenMultipleRetries_ShouldProcessOnlyDueOnes(t *testing.T) {
	setup := newRetryTestSetup(t)

	// Test scenario: mix of due and not-due retries
	notDue := createTestRetryInfo(1*time.Hour, 0, false)
	alreadyRetrying := createTestRetryInfo(-1*time.Hour, 1, true)

	setup.storeRetryInfo(t, "peer-1", notDue)
	setup.storeRetryInfo(t, "peer-2", alreadyRetrying)

	setup.coordinator.processSERetries(setup.ctx)

	// Verify neither was modified
	key1 := keys.NewPeerstoreSERetry("peer-1", notDue.CollectionID, notDue.DocID)
	key2 := keys.NewPeerstoreSERetry("peer-2", alreadyRetrying.CollectionID, alreadyRetrying.DocID)

	info1 := setup.getRetryInfo(t, key1)
	require.False(t, info1.Retrying, "Not-due retry should remain unchanged")
	require.Equal(t, 0, info1.NumRetries)

	info2 := setup.getRetryInfo(t, key2)
	require.True(t, info2.Retrying, "Already-retrying should remain unchanged")
	require.Equal(t, 1, info2.NumRetries)
}
