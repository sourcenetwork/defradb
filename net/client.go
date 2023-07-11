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
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/logging"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	PushTimeout = time.Second * 10
	PullTimeout = time.Second * 10
)

// pushLog creates a pushLog request and sends it to another node
// over libp2p grpc connection
func (s *server) pushLog(ctx context.Context, evt events.Update, pid peer.ID) error {
	log.Debug(
		ctx,
		"Preparing pushLog request",
		logging.NewKV("DocKey", evt.DocKey),
		logging.NewKV("CID", evt.Cid),
		logging.NewKV("SchemaId", evt.SchemaID))

	body := &pb.PushLogRequest_Body{
		DocKey:   []byte(evt.DocKey),
		Cid:      evt.Cid.Bytes(),
		SchemaID: []byte(evt.SchemaID),
		Creator:  s.peer.host.ID().String(),
		Log: &pb.Document_Log{
			Block: evt.Block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	log.Debug(
		ctx, "Pushing log",
		logging.NewKV("DocKey", evt.DocKey),
		logging.NewKV("CID", evt.Cid),
		logging.NewKV("PeerID", pid),
	)

	client, err := s.dial(pid) // grpc dial over p2p stream
	if err != nil {
		return NewErrPushLog(err)
	}

	cctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	if _, err := client.PushLog(cctx, req); err != nil {
		return NewErrPushLog(
			err,
			errors.NewKV("CID", evt.Cid),
			errors.NewKV("DocKey", evt.DocKey),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}
