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
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

// Log represents a new DAG node added to the
// append-only MerkleCRDT Clock graph of a
// document or sub-field.
// Note: This may need to be an interface :/
type Log struct {
	DocKey   string
	Cid      cid.Cid
	SchemaID string
	Block    ipld.Node
	Priority uint64
}
