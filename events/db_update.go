// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package events

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/sourcenetwork/immutable"
)

// UpdateChannel is the bus onto which updates are published.
type UpdateChannel = immutable.Option[Channel[Update]]

// EmptyUpdateChannel is an empty UpdateChannel.
var EmptyUpdateChannel = immutable.None[Channel[Update]]()

// UpdateEvent represents a new DAG node added to the append-only composite MerkleCRDT Clock graph
// of a document.
//
// It must only contain public elements not protected by ACP.
type Update struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Block is the contents of this composite commit, it contains the Cids of the field level commits that
	// also formed this update.
	Block ipld.Node

	// Priority is used to determine the order in which concurrent updates are applied.
	Priority uint64
}
