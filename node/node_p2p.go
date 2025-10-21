// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// P2P networking stack does not work in JS builds.
//
//go:build !js

package node

import (
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/go-p2p"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
)

func (n *Node) startP2P(ctx context.Context, store corekv.ReaderWriter, chunkSize immutable.Option[int]) error {
	if n.config.disableP2P {
		return nil
	}

	n.options = append(n.options, p2p.WithBlockstore(datastore.P2PBlockstoreFrom(store, chunkSize)))

	peer, err := p2p.NewPeer(
		ctx,
		filterOptions[p2p.NodeOpt](n.options)...,
	)
	if err != nil {
		return err
	}
	n.options = append(n.options, db.WithP2P(peer))
	n.peer = peer
	return nil
}
