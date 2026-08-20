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

var FieldCRDTs = []FieldValueCRDT{
	NewLWW(),
	NewCounter(true),
	NewCounter(false),
}

func TryGetFieldCRDT(ct client.CType) (FieldValueCRDT, bool) {
	if ct == client.NONE_CRDT {
		return nil, true
	}

	for _, crdt := range FieldCRDTs {
		if crdt.CType() == ct {
			return crdt, true
		}
	}
	return nil, false
}

type KindLimitedCRDT interface {
	SupportedKinds() []client.FieldKind
}

type FieldValueCRDT interface {
	CType() client.CType

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
