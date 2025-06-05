// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package admin

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/defradb/client"
)

func (adminDB *AdminDB) PeerInfo() peer.AddrInfo {
	// checkAdminAccess(ctx, adminDB.adminACP.Value(), acpTypes.AdminDACPerm)
	return adminDB.db.PeerInfo()
}

func (adminDB *AdminDB) SetReplicator(ctx context.Context, rep client.ReplicatorParams) error {
	//hasAccess, err := checkAdminAccess(ctx, adminDB.adminACP.Value(), acpTypes.AdminDACPerm)
	//if err != nil {
	//	return err
	//}

	//if !hasAccess {
	//	return errors.New("No access")
	//}

	return adminDB.db.SetReplicator(ctx, rep)

}

func (adminDB *AdminDB) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return adminDB.db.GetAllReplicators(ctx)

}

func (adminDB *AdminDB) DeleteReplicator(ctx context.Context, rep client.ReplicatorParams) error {
	return adminDB.db.DeleteReplicator(ctx, rep)

}

func (adminDB *AdminDB) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	return adminDB.db.AddP2PCollections(ctx, collectionIDs)
}

func (adminDB *AdminDB) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return adminDB.db.GetAllP2PCollections(ctx)
}

func (adminDB *AdminDB) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	return adminDB.db.RemoveP2PCollections(ctx, collectionIDs)
}
