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
	"encoding/binary"
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

// TestDecodeEnvelope_LegacyUvarint checks the fallback that reads a record written by the
// pre-envelope format: a bare uvarint status with no reason or payload.
func TestDecodeEnvelope_LegacyUvarint(t *testing.T) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(client.InProgressActionStatus))

	env, err := DecodeEnvelope(buf[:n])
	require.NoError(t, err)
	assert.Equal(t, client.InProgressActionStatus, env.Status)
	assert.Empty(t, env.Reason)
	assert.Empty(t, env.Payload)
}

// TestDecodeEnvelope_Corrupt checks a value that is neither valid JSON nor a valid uvarint
// surfaces an error rather than silently decoding to a zero status.
func TestDecodeEnvelope_Corrupt(t *testing.T) {
	// An empty value yields n<=0 from binary.Uvarint and is not valid JSON.
	_, err := DecodeEnvelope([]byte{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCorruptActionRecord), "expected ErrCorruptActionRecord, got: %v", err)
}

// TestRegisterSubject_SecondRegistrationWhileInProgress_Errors checks the guard that prevents
// two concurrent executions of the same action+subject on one collection.
func TestRegisterSubject_SecondRegistrationWhileInProgress_Errors(t *testing.T) {
	ctx := context.Background()
	ms, events := newTestMultistore(t)

	err := RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)

	err = RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrActionInProgress), "expected ErrActionInProgress, got: %v", err)
}

// TestRegisterSubject_DistinctSubjects_DoNotCollide checks two builds with different subjects
// (e.g. distinct index IDs) on the same collection both register independently.
func TestRegisterSubject_DistinctSubjects_DoNotCollide(t *testing.T) {
	ctx := context.Background()
	ms, events := newTestMultistore(t)

	require.NoError(t, RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1"))
	require.NoError(t, RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "2"))
}

// TestRegisterSubject_AfterComplete_Succeeds checks the guard releases once the prior
// execution completes (its record is deleted).
func TestRegisterSubject_AfterComplete_Succeeds(t *testing.T) {
	ctx := context.Background()
	ms, events := newTestMultistore(t)

	require.NoError(t, RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1"))
	require.NoError(t, CompleteSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1"))
	require.NoError(t, RegisterSubject(ctx, ms, events, "col1", client.BackfillIndexAction, "1"))
}
