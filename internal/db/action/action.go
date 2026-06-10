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
	"errors"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// Register a new action for execution
func Register(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
) error {
	status, err := getStatus(ctx, multistore, collectionID, action)
	if err != nil {
		return err
	}
	if status == client.InProgressActionStatus {
		return NewErrActionInProgress(collectionID, action)
	}

	return Set(ctx, multistore, events, collectionID, action, client.InProgressActionStatus)
}

// Set the status for an existing action
func Set(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
	status client.ActionStatus,
) error {
	val := make([]byte, binary.MaxVarintLen16)
	binary.PutUvarint(val, uint64(status))

	err := multistore.Systemstore().Set(
		ctx,
		keys.NewActionStatusKey(collectionID, action).Bytes(),
		val,
	)

	events.Publish(event.NewMessage(event.ActionExecutionName, event.ActionExecution{
		CollectionID: collectionID,
		Action:       action,
		Status:       status,
	}))

	return err
}

// Complete an existing action
func Complete(
	ctx context.Context,
	multistore *datastore.Multistore,
	events event.Bus,
	collectionID string,
	action client.Action,
) error {
	err := multistore.Systemstore().Delete(
		ctx,
		keys.NewActionStatusKey(collectionID, action).Bytes(),
	)

	events.Publish(event.NewMessage(event.ActionExecutionName, event.ActionExecution{
		CollectionID: collectionID,
		Action:       action,
		Status:       client.CompletedActionStatus,
	}))

	return err
}

func getStatus(
	ctx context.Context,
	multistore *datastore.Multistore,
	collectionID string,
	action client.Action,
) (client.ActionStatus, error) {
	val, err := multistore.Systemstore().Get(
		ctx,
		keys.NewActionStatusKey(collectionID, action).Bytes(),
	)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return client.NoneActionStatus, nil
		}
		return 0, err
	}

	intVal, _ := binary.Uvarint(val)

	return client.ActionStatus(intVal), nil
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

		intVal, _ := binary.Uvarint(val)
		status := client.ActionStatus(intVal)

		results = append(results, client.ActionExecution{
			CollectionID: key.CollectionID,
			Action:       key.Action,
			Status:       status,
		})
	}

	err = iter.Close()
	if err != nil {
		return nil, err
	}

	return results, nil
}
