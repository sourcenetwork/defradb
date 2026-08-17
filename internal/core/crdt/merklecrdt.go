// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package crdt provides CRDT implementations leveraging MerkleClock.
*/
package crdt

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type FieldLevelCRDT interface {
	ReplicatedData
	Delta(
		ctx context.Context,
		collectionVersionID string,
		data *DocField,
		isAdd bool,
	) (Delta, error)
}

func FieldLevelCRDTWithStore(
	cType client.CType,
	kind client.FieldKind,
	key keys.DataStoreKey,
) (FieldLevelCRDT, error) {
	switch cType {
	case client.LWW_REGISTER:
		return NewLWW(
			key,
		), nil
	case client.PN_COUNTER, client.P_COUNTER:
		return NewCounter(
			key,
			cType == client.PN_COUNTER,
			kind.(client.ScalarKind), //nolint:forcetypeassert
		), nil
	}
	return nil, client.NewErrUnknownCRDT(cType)
}
