// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"context"

	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/kms"
	"github.com/sourcenetwork/defradb/net"
)

func (n *Node) startP2P(ctx context.Context) error {
	if n.config.disableP2P {
		return nil
	}
	peer, err := net.NewPeer(
		ctx,
		n.DB.Events(),
		n.acp,
		n.DB.(*db.DB),
		filterOptions[net.NodeOpt](n.options)...,
	)
	if err != nil {
		return err
	}
	n.Peer = peer

	ident, err := n.DB.GetNodeIdentity(ctx)
	if err != nil {
		return err
	}
	if n.config.kmsType.HasValue() {
		switch n.config.kmsType.Value() {
		case kms.PubSubServiceType:
			n.kmsService, err = kms.NewPubSubService(
				ctx,
				peer.PeerID(),
				peer.Server(),
				n.DB.Events(),
				n.DB.Encstore(),
				n.acp,
				db.NewCollectionRetriever(n.DB),
				ident.Value().DID,
			)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
