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

// Envelope is the stored value of an action record. Reason holds an errored action's cause
// and Payload holds action-specific state such as an index build watermark. DecodeEnvelope
// also accepts the legacy bare-uvarint encoding.
type Envelope struct {
	Status  client.ActionStatus `json:"status"`
	Reason  string              `json:"reason,omitempty"`
	Payload json.RawMessage     `json:"payload,omitempty"`
}

// encodeValue serializes an action record value.
func encodeValue(status client.ActionStatus, reason string, payload json.RawMessage) ([]byte, error) {
	return json.Marshal(Envelope{Status: status, Reason: reason, Payload: payload})
}

// DecodeEnvelope deserializes an action record value, falling back to the legacy bare-uvarint
// encoding (a status with no reason or payload). A value that is neither valid JSON nor a
// valid uvarint is corrupt and returns an error rather than a silent zero status.
func DecodeEnvelope(val []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(val, &env); err == nil {
		return env, nil
	}
	intVal, n := binary.Uvarint(val)
	if n <= 0 {
		return Envelope{}, NewErrCorruptActionRecord(val)
	}
	return Envelope{Status: client.ActionStatus(intVal)}, nil
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
	return RegisterSubject(ctx, multistore, events, collectionID, action, "")
}

// RegisterSubject registers a new per-subject action for execution.
//
// Subject distinguishes concurrent executions of the same action on one collection
// (for example, an index build keyed by index ID). Pass an empty subject for
// collection-wide actions.
func RegisterSubject(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
) error {
	status, err := getStatus(multistore, collectionID, action, subject)
	if err != nil {
		return err
	}
	if status == client.InProgressActionStatus {
		return NewErrActionInProgress(collectionID, action)
	}

	return setSubject(
		ctx, multistore, events, collectionID, action, subject,
		client.InProgressActionStatus, "", nil,
	)
}

// Set the status for an existing action. Non-transactional, collection-wide.
func Set(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	status client.ActionStatus,
) error {
	return setSubject(ctx, multistore, events, collectionID, action, "", status, "", nil)
}

// setSubject writes an action record non-transactionally.
//
// It passes context.TODO() to force a transaction-free write: corekv otherwise binds the
// transaction on the context to the write (https://github.com/sourcenetwork/corekv/issues/107).
func setSubject(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
	status client.ActionStatus,
	reason string,
	payload json.RawMessage,
) error {
	val, err := encodeValue(status, reason, payload)
	if err != nil {
		return err
	}

	err = multistore.Systemstore().Set(
		context.TODO(),
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
		val,
	)
	if err != nil {
		return err
	}

	publish(events, collectionID, action, subject, status)
	return nil
}

// SetTxn writes a per-subject action record within the transaction bound to ctx, so the record
// commits atomically with other writes on it (for example an index build watermark with the
// index entries of the same batch).
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

	val, err := encodeValue(status, reason, payload)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Set(
		ctx,
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
		val,
	)
	if err != nil {
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

// Complete a collection-wide action by deleting its record. Non-transactional.
func Complete(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
) error {
	return CompleteSubject(ctx, multistore, events, collectionID, action, "")
}

// CompleteSubject completes a per-subject action by deleting its record. Non-transactional.
//
// A missing record means the action has completed, so completion deletes rather than
// storing a terminal status.
func CompleteSubject(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
) error {
	err := multistore.Systemstore().Delete(
		// See setSubject for why this is transaction-free.
		context.TODO(),
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
	)
	if err != nil {
		return err
	}

	publish(events, collectionID, action, subject, client.CompletedActionStatus)
	return nil
}

// CompleteTxn completes a per-subject action within the transaction bound to ctx, deleting
// its record atomically with other writes on the same transaction.
func CompleteTxn(
	ctx context.Context,
	events event.Bus,
	collectionID string,
	action client.Action,
	subject string,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	err := txn.Systemstore().Delete(
		ctx,
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
	)
	if err != nil {
		return err
	}

	txn.OnSuccess(func() {
		publish(events, collectionID, action, subject, client.CompletedActionStatus)
	})
	return nil
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
	subject string,
) (client.ActionStatus, error) {
	val, err := multistore.Systemstore().Get(
		// See setSubject for why this is transaction-free.
		context.TODO(),
		keys.NewActionStatusSubjectKey(collectionID, action, subject).Bytes(),
	)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return client.NoneActionStatus, nil
		}
		return 0, err
	}

	env, err := DecodeEnvelope(val)
	if err != nil {
		return 0, err
	}
	return env.Status, nil
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

		env, err := DecodeEnvelope(val)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		results = append(results, client.ActionExecution{
			CollectionID: key.CollectionID,
			Action:       key.Action,
			Subject:      key.Subject,
			Status:       env.Status,
		})
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	return results, nil
}
