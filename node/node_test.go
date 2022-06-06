// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"testing"

	"github.com/sourcenetwork/defradb/db"
)

// NewNode
// options are properly set
// Node is properly initialized
// with HostKey
// with a DB
// pubsub yes or no

// db fixture


// broadcaster fixture

func TestNewNode(t *testing.T) {
	// NewNode(context.Background(), )
	db.NewTestDB(t)
}

// Node.Boostrap
// properly fails if when a bunch of invalid peers are provided
// uses internal logger ??

// Node.Close
// verify .Peer is properly closed
// no error
// pubsub topics are closed
// grpc server is stopped
// broadcaster bus is discarded
// peer is canceled
