// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type dummyHost struct {
	SimpleMockHost
}

func (d *dummyHost) Connect(ctx context.Context, addresses []string) error {
	return nil
}

type dummyPushProtocol struct{}

func (d *dummyPushProtocol) SendRequest(
	ctx context.Context,
	req protocol.PushLogRequest,
	peer string,
) (protocol.PushLogReply, error) {
	return protocol.PushLogReply{}, nil
}

func TestPushLogToReplicators_ConcurrentMapAccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &P2P{
		ctx:                ctx,
		host:               &dummyHost{},
		replicators:        make(map[string]map[string][]string),
		replicatorProtocol: &dummyPushProtocol{},
	}

	colID := "col1"
	p.replicators[colID] = map[string][]string{
		"peer1": {"addr1"},
		"peer2": {"addr2"},
	}

	var wg sync.WaitGroup
	// Concurrently read and push logs
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p.pushLogToReplicators(event.Update{
				CollectionID: colID,
				DocID:        fmt.Sprintf("doc-%d", idx),
			})
		}(i)
	}

	// Concurrently mutate replicators
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p.updateReplicators(
				ctx,
				fmt.Sprintf("peer-%d", idx),
				[]string{fmt.Sprintf("addr-%d", idx)},
				map[string]struct{}{colID: {}},
			)
		}(i)
	}

	wg.Wait()
}

func TestHandleCompletedReplicatorRetry_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &P2P{}
	err := p.handleCompletedReplicatorRetry(ctx, "peer1", true)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBackpressureWait_OnlyPeerRejectionsAreTransient(t *testing.T) {
	_, transient := backpressureWait(errors.New(peerAtCapacityMessage))
	assert.False(t, transient, "a local error that happens to contain the text is not backpressure")

	wait, transient := backpressureWait(message.NewErrPeerRejected(peerRateLimitedMessage))
	assert.True(t, transient)
	assert.Equal(t, 100*time.Millisecond, wait)

	_, transient = backpressureWait(message.NewErrPeerRejected("unknown collection"))
	assert.False(t, transient)
}

func TestRetryWithBackpressure_StopsOnNonTransientError(t *testing.T) {
	calls := 0
	rejections := []error{message.NewErrPeerRejected(peerRateLimitedMessage), message.NewErrPeerRejected("bad block")}
	err := retryWithBackpressure(context.Background(), func() error {
		calls++
		return rejections[min(calls, len(rejections))-1]
	})
	assert.ErrorContains(t, err, "bad block")
	assert.Equal(t, 2, calls)
}

func TestRetryWithBackpressure_CancelledWhileWaiting_ReturnsLastError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := retryWithBackpressure(ctx, func() error { return message.NewErrPeerRejected(peerAtCapacityMessage) })
	assert.ErrorIs(t, err, message.ErrPeerRejected)
}

func TestRetryWithBackpressure_StopsWhenBudgetRunsOut(t *testing.T) {
	calls := 0
	start := time.Now()
	err := retryWithBackpressureFor(context.Background(), 250*time.Millisecond, func() error {
		calls++
		return message.NewErrPeerRejected(peerRateLimitedMessage)
	})
	assert.ErrorIs(t, err, message.ErrPeerRejected)
	assert.Less(t, time.Since(start), time.Second, "the budget, not the attempt cap, must end the loop")
	assert.LessOrEqual(t, calls, 4)
}

func TestAddReplicatorNextRetry_Restart_ReturnsToFirstInterval(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	intervals := []time.Duration{time.Minute, time.Hour}
	key := keys.NewReplicatorRetryIDKey("peer1")
	b, err := cbor.Marshal(retryInfo{NumRetries: 5, Retrying: true})
	require.NoError(t, err)
	require.NoError(t, store.Set(ctx, key.Bytes(), b))

	require.NoError(t, addReplicatorNextRetry(ctx, "peer1", intervals, true, store))

	b, err = store.Get(ctx, key.Bytes())
	require.NoError(t, err)
	var rInfo retryInfo
	require.NoError(t, cbor.Unmarshal(b, &rInfo))
	assert.Equal(t, 0, rInfo.NumRetries, "the next failing sweep must climb from the first interval again")
	assert.False(t, rInfo.Retrying)
	assert.WithinDuration(t, time.Now().Add(time.Minute), rInfo.NextRetry, 5*time.Second)
}

type multistoreDB struct {
	DB
	ms *datastore.Multistore
}

func (d *multistoreDB) Multistore() *datastore.Multistore { return d.ms }

func TestRetryReplicatorItems_Cancelled_ClearsRetrying(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ms := datastore.NewMultistore(memory.NewDatastore(context.Background()), nil, immutable.None[int]())
	key := keys.NewReplicatorRetryIDKey("peer1")
	b, err := cbor.Marshal(retryInfo{NumRetries: 3, Retrying: true})
	require.NoError(t, err)
	require.NoError(t, ms.Peerstore().Set(ctx, key.Bytes(), b))
	p := &P2P{db: &multistoreDB{ms: ms}}

	cancel()
	p.retryReplicatorItems(ctx, "peer1")

	b, err = ms.Peerstore().Get(context.Background(), key.Bytes())
	require.NoError(t, err)
	var rInfo retryInfo
	require.NoError(t, cbor.Unmarshal(b, &rInfo))
	assert.False(t, rInfo.Retrying, "a stale flag makes retryReplicators skip this replicator for good")
	assert.Equal(t, 3, rInfo.NumRetries)
}

// cancelledAfterLoopCtx reports cancellation through Err but never closes Done, standing in for
// a cancel that lands after the sweep loop's own check, such as during a backpressure wait.
type cancelledAfterLoopCtx struct {
	context.Context
}

func (cancelledAfterLoopCtx) Err() error { return context.Canceled }

func TestRetryReplicatorSweeps_CancelledDuringLoop_DoNotReportCompletion(t *testing.T) {
	ms := datastore.NewMultistore(memory.NewDatastore(context.Background()), nil, immutable.None[int]())
	p := &P2P{db: &multistoreDB{ms: ms}}
	ctx := cancelledAfterLoopCtx{context.Background()}

	ranToCompletion, _, _ := p.retryReplicatorDocs(ctx, "peer1")
	assert.False(t, ranToCompletion, "a cancelled sweep must clear its Retrying flag, not report completion")

	ranToCompletion, _, _ = p.retryReplicatorCols(ctx, "peer1")
	assert.False(t, ranToCompletion, "a cancelled sweep must clear its Retrying flag, not report completion")
}

func TestClearStaleReplicatorRetrying_ClearsOnlyTheFlag(t *testing.T) {
	ctx := context.Background()
	ms := datastore.NewMultistore(memory.NewDatastore(ctx), nil, immutable.None[int]())
	nextRetry := time.Now().Add(time.Hour)
	stuck := keys.NewReplicatorRetryIDKey("peer1")
	idle := keys.NewReplicatorRetryIDKey("peer2")
	for key, rInfo := range map[keys.ReplicatorRetryIDKey]retryInfo{
		stuck: {NextRetry: nextRetry, NumRetries: 3, Retrying: true},
		idle:  {NextRetry: nextRetry, NumRetries: 1},
	} {
		b, err := cbor.Marshal(rInfo)
		require.NoError(t, err)
		require.NoError(t, ms.Peerstore().Set(ctx, key.Bytes(), b))
	}
	p := &P2P{db: &multistoreDB{ms: ms}}

	p.clearStaleReplicatorRetrying(ctx)

	for key, wantRetries := range map[keys.ReplicatorRetryIDKey]int{stuck: 3, idle: 1} {
		b, err := ms.Peerstore().Get(ctx, key.Bytes())
		require.NoError(t, err)
		var rInfo retryInfo
		require.NoError(t, cbor.Unmarshal(b, &rInfo))
		assert.False(t, rInfo.Retrying, "no sweep runs before startup, so every flag left is stale")
		assert.Equal(t, wantRetries, rInfo.NumRetries)
		// cbor stores times to the second.
		assert.WithinDuration(t, nextRetry, rInfo.NextRetry, time.Second, "the next retry time is kept")
	}
}
