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
	"github.com/sourcenetwork/defradb/internal/se"
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
			handleRepErr := s.peer.handleReplicatorFailure(s.peer.ctx, pid.String(), evt.DocID)
			if handleRepErr != nil {
				err = errors.Join(err, handleRepErr)
			}
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

// pushSEArtifacts creates and sends SE artifacts to another node
func (s *server) pushSEArtifacts(evt se.ReplicateEvent, pid peer.ID) (err error) {
	defer func() {
		if err != nil && !evt.IsRetry {
			// Collect unique field names from artifacts
			fieldNamesMap := make(map[string]struct{})
			for _, artifact := range evt.Artifacts {
				fieldNamesMap[artifact.FieldName] = struct{}{}
			}

			var fieldNames []string
			for fieldName := range fieldNamesMap {
				fieldNames = append(fieldNames, fieldName)
			}

			s.peer.bus.Publish(event.NewMessage(se.ReplicationFailureEventName, se.ReplicationFailureEvent{
				DocID:        evt.DocID,
				CollectionID: evt.CollectionID,
				PeerID:       pid,
				FieldNames:   fieldNames,
			}))
		}
		if evt.Success != nil {
			evt.Success <- err == nil
		}
	}()

	client, err := s.dial(pid)
	if err != nil {
		return NewErrPushSEArtifacts(err)
	}

	ctx, cancel := context.WithTimeout(s.peer.ctx, PushTimeout)
	defer cancel()

	netArtifacts := make([]seArtifact, len(evt.Artifacts))
	for i, artifact := range evt.Artifacts {
		netArtifacts[i] = seArtifact{
			DocID:     artifact.DocID,
			IndexID:   artifact.IndexID,
			SearchTag: artifact.SearchTag,
		}
	}

	req := pushSEArtifactsRequest{
		CollectionID: evt.CollectionID,
		Artifacts:    netArtifacts,
		Creator:      s.peer.host.ID().String(),
	}

	if err := client.Invoke(ctx, servicePushSEArtifactsName, req, nil); err != nil {
		return NewErrPushSEArtifacts(
			err,
			errors.NewKV("DocID", evt.DocID),
			errors.NewKV("CollectionID", evt.CollectionID),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}

// querySEArtifacts queries SE artifacts on a remote node
func (s *server) querySEArtifacts(ctx context.Context, pid peer.ID, req querySEArtifactsRequest) (*querySEArtifactsReply, error) {
	client, err := s.dial(pid)
	if err != nil {
		return nil, NewErrQuerySEArtifacts(err)
	}

	ctx, cancel := context.WithTimeout(ctx, PullTimeout)
	defer cancel()

	resp := &querySEArtifactsReply{}
	if err := client.Invoke(ctx, serviceQuerySEArtifactsName, req, resp); err != nil {
		return nil, NewErrQuerySEArtifacts(err,
			errors.NewKV("CollectionID", req.CollectionID),
			errors.NewKV("PeerID", pid),
		)
	}

	return resp, nil
}
