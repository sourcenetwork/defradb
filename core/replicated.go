// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"context"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

// ReplicatedData is a data type that allows concurrent writers to deterministically merge other
// replicated data so as to converge on the same state.
type ReplicatedData interface {
	ID() string
	Merge(ctx context.Context, other Delta) error
	DeltaDecode(node ipld.Node) (Delta, error) // possibly rename to just Decode
	Value(ctx context.Context) ([]byte, error)
}

// PersistedReplicatedData persists a ReplicatedData to an underlying datastore.
type PersistedReplicatedData interface {
	ReplicatedData
	Publish(Delta) (cid.Cid, error)
}

// type EmbedableReplicatedData interface {
// 	ReplicatedData
// 	Apply(Operation) error
// }
