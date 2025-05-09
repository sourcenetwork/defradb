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
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	PushTimeout = time.Second * 10
	PullTimeout = time.Second * 10
)

// pushLog creates a pushLog request and sends it to another node
// over libp2p grpc connection
func (s *server) pushLog(evt event.Update, pid peer.ID) (err error) {
	defer func() {
		// When the event is a retry, we don't need to republish the failure as
		// it is already being handled by the retry mechanism through the success channel.
		if err != nil && !evt.IsRetry {
			s.peer.bus.Publish(event.NewMessage(event.ReplicatorFailureName, event.ReplicatorFailure{
				DocID:  evt.DocID,
				PeerID: pid,
			}))
		}
		// Success is not nil when the pushLog is called from a retry
		if evt.Success != nil {
			evt.Success <- err == nil
		}
	}()

	client, err := s.dial(pid) // grpc dial over P2P stream
	if err != nil {
		return NewErrPushLog(err)
	}

	ctx, cancel := context.WithTimeout(s.peer.ctx, PushTimeout)
	defer cancel()

	req := pushLogRequest{
		DocID:        evt.DocID,
		CID:          evt.Cid.Bytes(),
		CollectionID: evt.CollectionID,
		Creator:      s.peer.host.ID().String(),
		Block:        evt.Block,
	}
	if err := client.Invoke(ctx, servicePushLogName, req, nil); err != nil {
		return NewErrPushLog(
			err,
			errors.NewKV("CID", evt.Cid),
			errors.NewKV("DocID", evt.DocID),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}

// getIdentity creates a getIdentity request and sends it to another node
func (s *server) getIdentity(ctx context.Context, pid peer.ID) (getIdentityReply, error) {
	client, err := s.dial(pid) // grpc dial over P2P stream
	if err != nil {
		return getIdentityReply{}, NewErrPushLog(err)
	}

	ctx, cancel := context.WithTimeout(ctx, PushTimeout)
	defer cancel()

	req := getIdentityRequest{
		PeerID: s.peer.host.ID().String(),
	}
	resp := getIdentityReply{}
	if err := client.Invoke(ctx, serviceGetIdentityName, req, &resp); err != nil {
		return getIdentityReply{}, NewErrFailedToGetIdentity(
			err,
			errors.NewKV("PeerID", pid),
		)
	}
	return resp, nil
}
