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
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type FieldValueCRDT interface {
	Merge(
		ctx context.Context,
		store datastore.Keyedstore,
		key keys.DataStoreKey,
		kind client.FieldKind,
		other Delta,
	) error
}

type DocumentValueCRDT interface {
	Merge(ctx context.Context, store datastore.Keyedstore, key keys.PrimaryDataStoreKey, other Delta) error
}

func FieldLevelCRDTWithStore(
	cType client.CType,
) (FieldValueCRDT, error) {
	switch cType {
	case client.LWW_REGISTER:
		return NewLWW(), nil
	case client.PN_COUNTER, client.P_COUNTER:
		return NewCounter(
			cType == client.PN_COUNTER,
		), nil
	}
	return nil, client.NewErrUnknownCRDT(cType)
}
