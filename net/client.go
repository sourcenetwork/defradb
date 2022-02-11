// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/sourcenetwork/defradb/net/pb"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document/key"
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	PushTimeout = time.Second * 10
	PullTimeout = time.Second * 10
)

// pushLog creates a pushLog request and sends it to another node
// over libp2p grpc connection
func (s *server) pushLog(ctx context.Context, lg core.Log, pid peer.ID) error {
	dockey, err := key.NewFromString(lg.DocKey)
	if err != nil {
		return fmt.Errorf("Failed to get DocKey from broadcast message: %w", err)
	}
	log.Debugf("Preparing pushLog request for rpc %s at %s using %s", dockey, lg.Cid, lg.SchemaID)
	body := &pb.PushLogRequest_Body{
		DocKey:   &pb.ProtoDocKey{DocKey: dockey},
		Cid:      &pb.ProtoCid{Cid: lg.Cid},
		SchemaID: []byte(lg.SchemaID),
		Log: &pb.Document_Log{
			Block: lg.Block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	log.Debugf("pushing log %s for %s to %s", lg.Cid, dockey, pid)
	client, err := s.dial(pid) // grpc dial over p2p stream
	if err != nil {
		return fmt.Errorf("Failed to push log: %w", err)
	}

	cctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	if _, err := client.PushLog(cctx, req); err != nil {
		return fmt.Errorf("Failed PushLog RPC request %s for %s to %s: %w", lg.Cid, dockey, pid, err)
	}
	return nil
}
