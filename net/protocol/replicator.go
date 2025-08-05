// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package protocol

import (
	"context"
	"errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/net/message"
)

// PushLogRequest is the struct used to send a resource update to a peer node
type PushLogRequest struct {
	message.MetaData
	DocID        string
	CID          []byte
	CollectionID string
	Creator      string
	Block        []byte
}

// PushLogReply is the expected response struct that should be received after
// an pushlog request.
type PushLogReply struct {
	message.MetaData
}

const (
	replicatorProtocolRequest  = "/defradb/rep_req/0.0.1"
	replicatorProtocolResponse = "/defradb/rep_resp/0.0.1"
)

type pushLogFunc func(
	ctx context.Context,
	req *PushLogRequest,
	isReplicator bool,
) (*PushLogReply, error)

type replicatorFailureFunc func(ctx context.Context, peerID, docID string) error

// ReplicatorProtocol is the protocol implementation for sending resource updates to a peer node.
type ReplicatorProtocol struct {
	*baseProto
	pushLogFunc           pushLogFunc
	replicatorFailureFunc replicatorFailureFunc
}

// NewReplicatorProtocol returns and a new [ReplicatorProtocol] struct and registers the protocol
// on the stream handler.
func NewReplicatorProtocol(
	h host.Host,
	pushLogFunc pushLogFunc,
	replicatorFailureFunc replicatorFailureFunc,
) *ReplicatorProtocol {
	proto := &ReplicatorProtocol{
		baseProto:             newBaseProto(h),
		pushLogFunc:           pushLogFunc,
		replicatorFailureFunc: replicatorFailureFunc,
	}
	h.SetStreamHandler(replicatorProtocolRequest, proto.onRequest)
	h.SetStreamHandler(replicatorProtocolResponse, proto.onResponse)
	return proto
}

// PushToReplicator sends the pushlog request to the provided peer node.
//
// Callers should set an appropriate context timeout.
func (proto *ReplicatorProtocol) PushToReplicator(
	ctx context.Context,
	evt event.Update,
	pid peer.ID,
) (reply *PushLogReply, err error) {
	defer func() {
		// When the event is a retry, we don't need to republish the failure as
		// it is already being handled by the retry mechanism through the success channel.
		if err != nil && !evt.IsRetry {
			handleRepErr := proto.replicatorFailureFunc(ctx, pid.String(), evt.DocID)
			if handleRepErr != nil {
				err = errors.Join(err, handleRepErr)
			}
		}
	}()

	req := PushLogRequest{
		DocID:        evt.DocID,
		CID:          evt.Cid.Bytes(),
		CollectionID: evt.CollectionID,
		Creator:      proto.host.ID().String(),
		Block:        evt.Block,
	}
	m, err := message.Send(ctx, proto, &req, pid, replicatorProtocolRequest, true)
	if err != nil {
		return nil, err
	}
	return m.(*PushLogReply), nil //nolint:forcetypeassert
}

func (proto *ReplicatorProtocol) onRequest(s network.Stream) {
	ctx := context.Background()
	var err error

	req := PushLogRequest{}
	err = message.Receive(s, proto, &req)
	if err != nil {
		return
	}

	defer func() {
		// if an error occurs, try to tell the node that sent the request what went wrong.
		if err != nil {
			resp := PushLogReply{}
			resp.SetMessageID(req.MessageID)
			resp.SetErrMessage(err.Error())
			_, _ = message.Send(ctx, proto, &resp, s.Conn().RemotePeer(), replicatorProtocolResponse, false)
		}
	}()

	resp, err := proto.pushLogFunc(ctx, &req, true)
	if err != nil {
		return
	}

	resp.SetMessageID(req.MessageID)
	_, _ = message.Send(ctx, proto, resp, s.Conn().RemotePeer(), replicatorProtocolResponse, false)
}

func (proto *ReplicatorProtocol) onResponse(s network.Stream) {
	resp := PushLogReply{}
	err := message.Receive(s, proto, &resp)
	if err != nil {
		return
	}
}
