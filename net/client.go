// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
