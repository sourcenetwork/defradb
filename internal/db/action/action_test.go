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
	"github.com/sourcenetwork/defradb/internal/keys"
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

// actionTxnEnv bundles the event bus and a transaction-bound context used by the SetTxn /
// CompleteTxn / GetPayload tests, which require a transaction on the context.
type actionTxnEnv struct {
	events event.Bus
}

func newActionTxnCtx(t *testing.T) (actionTxnEnv, context.Context, func()) {
	t.Helper()
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()
	txn := datastore.NewTxnFrom(rootstore, lockSet, 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)

	events := event.NewChannelBus(0, 0)
	t.Cleanup(events.Close)

	return actionTxnEnv{events: events}, ctx, func() { txn.Discard() }
}

// TestStatus_RoundTrips checks the status value encodes and decodes as a bare uvarint.
func TestStatus_RoundTrips(t *testing.T) {
	for _, status := range []client.ActionStatus{
		client.InProgressActionStatus,
		client.ErroredActionStatus,
		client.CompletedActionStatus,
	} {
		assert.Equal(t, status, DecodeStatus(encodeStatus(status)))
	}
}

// TestSetTxn_StoresStatusReasonAndPayloadSeparately checks status, reason and payload live
// under distinct keys: the status is readable as a bare uvarint without touching the others,
// and the reason and opaque payload are independently retrievable.
func TestSetTxn_StoresStatusReasonAndPayloadSeparately(t *testing.T) {
	env, ctx, cleanup := newActionTxnCtx(t)
	defer cleanup()

	require.NoError(t, SetTxn(ctx, env.events, "col1", client.BackfillIndexAction, "1",
		client.ErroredActionStatus, "boom", json.RawMessage(`{"watermark":"w1"}`), true))

	txn := datastore.CtxMustGetTxn(ctx)

	statusVal, err := txn.Systemstore().Get(ctx, keys.NewActionStatusSubjectKey("col1", client.BackfillIndexAction, "1").Bytes())
	require.NoError(t, err)
	assert.Equal(t, client.ErroredActionStatus, DecodeStatus(statusVal))

	reason, err := GetReason(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.Equal(t, "boom", reason)

	payload, err := GetPayload(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"watermark":"w1"}`, string(payload))
}

// TestSetTxn_EmptyClearsPriorReasonAndPayload checks that updating with an empty reason and
// payload removes values written by a previous status.
func TestSetTxn_EmptyClearsPriorReasonAndPayload(t *testing.T) {
	env, ctx, cleanup := newActionTxnCtx(t)
	defer cleanup()

	require.NoError(t, SetTxn(ctx, env.events, "col1", client.BackfillIndexAction, "1",
		client.ErroredActionStatus, "boom", json.RawMessage(`{"watermark":"w1"}`), false))
	require.NoError(t, SetTxn(ctx, env.events, "col1", client.BackfillIndexAction, "1",
		client.InProgressActionStatus, "", nil, false))

	reason, err := GetReason(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.Empty(t, reason)

	payload, err := GetPayload(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.Nil(t, payload)
}

// TestGetReasonAndPayload_Missing checks that missing reason/payload records are not errors.
func TestGetReasonAndPayload_Missing(t *testing.T) {
	_, ctx, cleanup := newActionTxnCtx(t)
	defer cleanup()

	reason, err := GetReason(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.Empty(t, reason)

	payload, err := GetPayload(ctx, "col1", client.BackfillIndexAction, "1")
	require.NoError(t, err)
	assert.Nil(t, payload)
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
