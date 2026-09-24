// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/wire"
)

// The IPLD blocks exchanged between nodes are wire types. Their shape is the IPLD
// schema each one declares, not the Go struct, so the snapshot records that
// schema. Registering them here puts them under the same wire changes check as the CBOR
// message types.
func init() {
	wire.Register[Block]()
	wire.Register[DAGLink]()
	wire.Register[Encryption]()
	wire.Register[Signature]()
	wire.Register[SignatureHeader]()

	wire.Register[crdt.CRDT]()
	wire.Register[crdt.LWWDelta]()
	wire.Register[crdt.DocCompositeDelta]()
	wire.Register[crdt.CounterDelta]()
	wire.Register[crdt.CollectionDelta]()
	wire.Register[crdt.CollectionSetDelta]()
	wire.Register[crdt.CollectionDefinitionDelta]()
	wire.Register[crdt.FieldDefinitionDelta]()
}
