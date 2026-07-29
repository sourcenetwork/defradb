// Copyright 2025 Democratized Data Foundation
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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

type SimpleMockHost struct {
	client.Host
}

func (m *SimpleMockHost) ID() string {
	return "peerID"
}

type mergeDB struct {
	DB
	active    atomic.Int32
	maxActive atomic.Int32
}

func (db *mergeDB) Merge(context.Context, event.Merge) error {
	active := db.active.Add(1)
	defer db.active.Add(-1)

	for {
		maxActive := db.maxActive.Load()
		if active <= maxActive || db.maxActive.CompareAndSwap(maxActive, active) {
			break
		}
	}
	time.Sleep(time.Millisecond)
	return nil
}

func TestMergeSerializesDatabaseWrites(t *testing.T) {
	db := &mergeDB{}
	p := &P2P{db: db}

	start := make(chan struct{})
	errs := make(chan error, 20)
	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			<-start
			errs <- p.merge(context.Background(), event.Merge{})
		})
	}
	close(start)
	group.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), db.maxActive.Load())
}

func TestPubSubMessageHandler_ContextCanceled(t *testing.T) {
	// Setup P2P with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	p := &P2P{
		ctx:  ctx, // This should trigger the early exit in processPushlogRequest
		host: &SimpleMockHost{},
	}

	// Create a dummy request message
	req := protocol.PushLogRequest{
		DocID: "docID",
		// Block can be empty or garbage, as context check should happen first
	}
	msg, err := cbor.Marshal(req)
	assert.NoError(t, err)

	// Call handler
	// from="sender", topic="topic"
	resp, err := p.pubSubMessageHandler("sender", "topic", msg)

	// Expectation: No error returned (suppressed), resp is nil
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestPubSubMessageHandler_ContextTimeout(t *testing.T) {
	// Setup P2P with timed out context
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	// Wait for context to be done
	<-ctx.Done()

	p := &P2P{
		ctx:  ctx,
		host: &SimpleMockHost{},
	}

	req := protocol.PushLogRequest{
		DocID: "docID",
	}
	msg, err := cbor.Marshal(req)
	assert.NoError(t, err)

	resp, err := p.pubSubMessageHandler("sender", "topic", msg)

	assert.NoError(t, err)
	assert.Nil(t, resp)
}
