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
	"io"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
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
	replicatorProtocolVersion  = "0.0.1"
	replicatorProtocolRequest  = "/defradb/rep_req/" + replicatorProtocolVersion
	replicatorProtocolResponse = "/defradb/rep_resp/" + replicatorProtocolVersion
)

type pushLogProcessorFunc func(
	ctx context.Context,
	req *PushLogRequest,
	isReplicator bool,
) error

type replicatorFailureFunc func(ctx context.Context, peerID, docID string) error

// ReplicatorProtocol is the protocol implementation for sending resource updates to a peer node.
type ReplicatorProtocol struct {
	*baseProto
	pushLogProcessorFunc  pushLogProcessorFunc
	replicatorFailureFunc replicatorFailureFunc
}

// NewReplicatorProtocol returns and a new [ReplicatorProtocol] struct and registers the protocol
// on the stream handler.
func NewReplicatorProtocol(
	h client.Host,
	pushLogProcessorFunc pushLogProcessorFunc,
	replicatorFailureFunc replicatorFailureFunc,
) *ReplicatorProtocol {
	proto := &ReplicatorProtocol{
		baseProto:             newBaseProto(h),
		pushLogProcessorFunc:  pushLogProcessorFunc,
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
	pid string,
) (reply *PushLogReply, err error) {
	defer func() {
		// When the event is a retry, we don't need to republish the failure as
		// it is already being handled by the retry mechanism through the success channel.
		if err != nil && !evt.IsRetry {
			handleRepErr := proto.replicatorFailureFunc(ctx, pid, evt.DocID)
			if handleRepErr != nil {
				err = errors.Join(err, handleRepErr)
			}
		}
	}()

	req := PushLogRequest{
		DocID:        evt.DocID,
		CID:          evt.Cid.Bytes(),
		CollectionID: evt.CollectionID,
		Creator:      proto.host.ID(),
		Block:        evt.Block,
	}
	return message.Send[*PushLogReply](ctx, proto, &req, pid, replicatorProtocolRequest)
}

func (proto *ReplicatorProtocol) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	req := PushLogRequest{}
	err := message.Receive(stream, peerID, proto, &req)
	if err != nil {
		return
	}

	defer func() {
		// if an error occurs, try to tell the node that sent the request what went wrong.
		if err != nil {
			resp := PushLogReply{}
			resp.SetMessageID(req.MessageID)
			resp.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, proto, &resp, peerID, replicatorProtocolResponse)
		}
	}()

	err = proto.pushLogProcessorFunc(ctx, &req, true)
	if err != nil {
		return
	}

	resp := PushLogReply{}
	resp.SetMessageID(req.MessageID)
	err = message.SendAndForget(ctx, proto, &resp, peerID, replicatorProtocolResponse)
}

func (proto *ReplicatorProtocol) onResponse(stream io.Reader, peerID string) {
	_ = message.Receive(stream, peerID, proto, &PushLogReply{})
}
