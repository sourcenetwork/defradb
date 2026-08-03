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
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

type SimpleMockHost struct {
	client.Host
}

func (m *SimpleMockHost) ID() string {
	return "peerID"
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
