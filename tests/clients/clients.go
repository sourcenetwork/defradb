// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clients

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

// Client implements the DB interface along with a few other methods
// required for testing.
type Client interface {
	client.DB
	client.P2P
	Connect(ctx context.Context, addr peer.AddrInfo) error
	Close()
	MaxTxnRetries() int
	Events() *event.Bus
}
