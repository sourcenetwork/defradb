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
	"testing"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
)

func newTestMultistore(t *testing.T) (*datastore.Multistore, event.Bus) {
	t.Helper()
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	ms := datastore.NewMultistore(rootstore, lock.NewLockSet(), immutable.None[int]())
	events := event.NewChannelBus(0, 0)
	t.Cleanup(events.Close)
	return ms, events
}

// TestDecodeEnvelope_JSON checks an envelope round-trips through the JSON encoding,
// preserving status, reason and payload.
func TestDecodeEnvelope_JSON(t *testing.T) {
	val, err := encodeValue(client.ErroredActionStatus, "boom", []byte(`{"watermark":"w1"}`))
	require.NoError(t, err)

	env, err := DecodeEnvelope(val)
	require.NoError(t, err)
	assert.Equal(t, client.ErroredActionStatus, env.Status)
	assert.Equal(t, "boom", env.Reason)
	assert.JSONEq(t, `{"watermark":"w1"}`, string(env.Payload))
}

// TestDecodeEnvelope_Corrupt checks a value that is not valid JSON surfaces an error rather
// than silently decoding to a zero status.
func TestDecodeEnvelope_Corrupt(t *testing.T) {
	_, err := DecodeEnvelope([]byte{0x00, 0x01, 0x02})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCorruptActionRecord), "expected ErrCorruptActionRecord, got: %v", err)
}

// TestRegister_SecondRegistrationWhileInProgress_Errors checks the guard that prevents two
// concurrent executions of the same action on one collection.
func TestRegister_SecondRegistrationWhileInProgress_Errors(t *testing.T) {
	ctx := context.Background()
	ms, events := newTestMultistore(t)

	err := Register(ctx, ms, events, "col1", client.TruncateAction)
	require.NoError(t, err)

	err = Register(ctx, ms, events, "col1", client.TruncateAction)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrActionInProgress), "expected ErrActionInProgress, got: %v", err)
}

// TestRegister_AfterComplete_Succeeds checks the guard releases once the prior execution
// completes (its record is deleted).
func TestRegister_AfterComplete_Succeeds(t *testing.T) {
	ctx := context.Background()
	ms, events := newTestMultistore(t)

	require.NoError(t, Register(ctx, ms, events, "col1", client.TruncateAction))
	require.NoError(t, Complete(ctx, ms, events, "col1", client.TruncateAction))
	require.NoError(t, Register(ctx, ms, events, "col1", client.TruncateAction))
}
