// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/events"
)

type UpdateChannel = Option[events.Channel[UpdateEvent]]

var EmptyUpdateChannel = None[events.Channel[UpdateEvent]]()

type Events struct {
	// Updates publishes an `UpdateEvent` for each document written to in the database.
	Updates UpdateChannel
}

// UpdateEvent represents a new DAG node added to the
// append-only MerkleCRDT Clock graph of a
// document or sub-field.
type UpdateEvent struct {
	DocKey   string
	Cid      cid.Cid
	SchemaID string
	Block    ipld.Node
	Priority uint64
}
