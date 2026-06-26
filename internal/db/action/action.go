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
	"encoding/json"
	"errors"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func encodeStatus(status client.ActionStatus) []byte {
	val := make([]byte, binary.MaxVarintLen16)
	n := binary.PutUvarint(val, uint64(status))
	return val[:n]
}

// DecodeStatus decodes an action status value (a bare uvarint).
func DecodeStatus(val []byte) client.ActionStatus {
	status, _ := binary.Uvarint(val)
	return client.ActionStatus(status)
}

// Register a new action for execution.
//
// The write is non-transactional. Use this for collection-wide actions whose work is
// not performed within a single transaction (truncate, datastore refresh).
func Register(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
) error {
	status, err := getStatus(multistore, collectionID, action)
	if err != nil {
		return err
	}
	if status == client.InProgressActionStatus {
		return NewErrActionInProgress(collectionID, action)
	}

	return Set(ctx, multistore, events, collectionID, action, client.InProgressActionStatus)
}

// Set the status for an existing action. Non-transactional, collection-wide.
//
// It passes context.TODO() to force a transaction-free write: corekv otherwise binds the
// transaction on the context to the write (https://github.com/sourcenetwork/corekv/issues/107).
func Set(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	status client.ActionStatus,
) error {
	err := multistore.Systemstore().Set(
		context.TODO(),
		keys.NewActionStatusKey(collectionID, action).Bytes(),
		encodeStatus(status),
	)
	if err != nil {
		return err
	}

	publish(events, collectionID, action, "", status)
	return nil
}

// SetTxn writes a per-subject action's status within the transaction bound to ctx, so the
// record commits atomically with other writes on it — for example an index build watermark
// with the index entries of the same batch.
//
// reason is the generic error reason (empty unless errored). payload is action-specific opaque
// data the caller encodes itself (nil when none). Each is stored under its own key, or cleared
// when empty.
//
// publishEvent reports whether an ActionExecution event is published on commit. Pass false for
// progress-only updates that repeat an already published status, to avoid one event per batch.
func SetTxn(
	ctx context.Context,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
	status client.ActionStatus,
	reason string,
	payload json.RawMessage,
	publishEvent bool,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	if err := txn.Systemstore().Set(
		ctx,
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
		encodeStatus(status),
	); err != nil {
		return err
	}

	if err := setOrClear(
		ctx, txn, keys.NewActionReasonKey(collectionID, action, subject).Bytes(), []byte(reason), reason != "",
	); err != nil {
		return err
	}
	if err := setOrClear(
		ctx, txn, keys.NewActionPayloadKey(collectionID, action, subject).Bytes(), payload, len(payload) > 0,
	); err != nil {
		return err
	}

	if publishEvent {
		// Publish on commit so listeners are not told about a record that may roll back.
		txn.OnSuccess(func() {
			publish(events, collectionID, action, subject, status)
		})
	}
	return nil
}

// setOrClear writes value at key when present is true, otherwise deletes any existing value.
func setOrClear(ctx context.Context, txn datastore.Txn, key, value []byte, present bool) error {
	if present {
		return txn.Systemstore().Set(ctx, key, value)
	}
	return txn.Systemstore().Delete(ctx, key)
}

// Complete a collection-wide action by deleting its record. Non-transactional.
//
// A missing record means the action has completed, so completion deletes rather than
// storing a terminal status.
func Complete(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
) error {
	err := multistore.Systemstore().Delete(
		// See Set for why this is transaction-free.
		context.TODO(),
		keys.NewActionStatusKey(collectionID, action).Bytes(),
	)
	if err != nil {
		return err
	}

	publish(events, collectionID, action, "", client.CompletedActionStatus)
	return nil
}

// CompleteTxn completes a per-subject action within the transaction bound to ctx, deleting its
// status, reason and payload records atomically with other writes on the same transaction.
func CompleteTxn(
	ctx context.Context,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
) error {
	if err := ClearTxn(ctx, collectionID, action, subject); err != nil {
		return err
	}

	txn := datastore.CtxMustGetTxn(ctx)
	txn.OnSuccess(func() {
		publish(events, collectionID, action, subject, client.CompletedActionStatus)
	})
	return nil
}

// ClearTxn deletes a per-subject action's status, reason and payload records on the transaction
// bound to ctx, without publishing an event. Use it when the record should vanish rather than
// report completion, e.g. clearing a failed build's record when its index is dropped, where a
// Completed event would wrongly signal success. Deleting a missing record is a no-op.
func ClearTxn(ctx context.Context, collectionID string, action client.Action, subject string) error {
	txn := datastore.CtxMustGetTxn(ctx)

	for _, key := range [][]byte{
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
		keys.NewActionReasonKey(collectionID, action, subject).Bytes(),
		keys.NewActionPayloadKey(collectionID, action, subject).Bytes(),
	} {
		if err := txn.Systemstore().Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// GetReason returns the error reason recorded for the given action execution, or "" if none.
func GetReason(ctx context.Context, collectionID string, action client.Action, subject string) (string, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	val, err := txn.Systemstore().Get(ctx, keys.NewActionReasonKey(collectionID, action, subject).Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return string(val), nil
}

// GetPayload returns the opaque payload bytes stored for the given action execution, or nil if
// none. The caller is responsible for decoding them.
func GetPayload(
	ctx context.Context,
	collectionID string,
	action client.Action,
	subject string,
) (json.RawMessage, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	val, err := txn.Systemstore().Get(ctx, keys.NewActionPayloadKey(collectionID, action, subject).Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return json.RawMessage(val), nil
}

func publish(
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
	status client.ActionStatus,
) {
	events.Publish(event.NewMessage(event.ActionExecutionName, event.ActionExecution{
		CollectionID: collectionID,
		Action:       action,
		Subject:      subject,
		Status:       status,
	}))
}

func getStatus(
	multistore *datastore.Multistore,
	collectionID string,
	action client.Action,
) (client.ActionStatus, error) {
	val, err := multistore.Systemstore().Get(
		// See Set for why this is transaction-free.
		context.TODO(),
		keys.NewActionStatusKey(collectionID, action).Bytes(),
	)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return client.NoneActionStatus, nil
		}
		return 0, err
	}

	return DecodeStatus(val), nil
}

// ListExecutions lists all the actions that have not yet successfully completed.
//
// This includes actions that failed, or have been cancelled.
func ListExecutions(ctx context.Context) ([]client.ActionExecution, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewEmptyActionStatusKey().Bytes(),
	})
	if err != nil {
		return nil, err
	}

	results := []client.ActionExecution{}
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		val, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		key, err := keys.NewActionStatusKeyString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		reason, err := GetReason(ctx, key.CollectionID, key.Action, key.Subject)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		results = append(results, client.ActionExecution{
			CollectionID: key.CollectionID,
			Action:       key.Action,
			Subject:      key.Subject,
			Status:       DecodeStatus(val),
			Reason:       reason,
		})
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	return results, nil
}
