// Copyright 2024 Democratized Data Foundation
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
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/immutable"
)

// DAGMergeChannel is the bus onto which dag merge are published.
type DAGMergeChannel = immutable.Option[Channel[DAGMerge]]

// DAGMerge is a notification that a merge can be performed up to the provided CID.
type DAGMerge struct {
	// DocID is the unique identifier for the document being merged.
	DocID string
	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid
	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string
	// Wg is a wait group that can be used to synchronize the merge,
	// allowing the caller to optionnaly block until the merge is complete.
	Wg *sync.WaitGroup
}
