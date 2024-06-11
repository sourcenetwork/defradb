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
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	PushTimeout = time.Second * 10
	PullTimeout = time.Second * 10
)

// pushLog creates a pushLog request and sends it to another node
// over libp2p grpc connection
func (s *server) pushLog(ctx context.Context, evt event.UpdateEvent, pid peer.ID) error {
	body := &pb.PushLogRequest_Body{
		DocID:      []byte(evt.DocID),
		Cid:        evt.Cid.Bytes(),
		SchemaRoot: []byte(evt.SchemaRoot),
		Creator:    s.peer.host.ID().String(),
		Log: &pb.Document_Log{
			Block: evt.Block,
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	client, err := s.dial(pid) // grpc dial over P2P stream
	if err != nil {
		return NewErrPushLog(err)
	}

	cctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	if _, err := client.PushLog(cctx, req); err != nil {
		return NewErrPushLog(
			err,
			errors.NewKV("CID", evt.Cid),
			errors.NewKV("DocID", evt.DocID),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}
