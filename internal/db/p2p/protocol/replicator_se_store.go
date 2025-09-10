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
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

// PushSEArtifactsRequest - Request to push SE artifacts
type PushSEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Artifacts    []SEArtifact
}

// SEArtifact - Network representation
type SEArtifact struct {
	DocID     string
	IndexID   string
	SearchTag []byte
}

// Reply type
type PushSEArtifactsReply struct {
	message.MetaData
}

const (
	replicatorSeProtocolVersion  = "0.0.1"
	replicatorSeProtocolRequest  = "/defradb/rep_se_req/" + replicatorSeProtocolVersion
	replicatorSeProtocolResponse = "/defradb/rep_se_resp/" + replicatorSeProtocolVersion
)

type pushSEArtifactsProcessorFunc func(
	ctx context.Context,
	req *PushSEArtifactsRequest,
	isReplicator bool,
) error

type replicatorSEFailureFunc func(ctx context.Context, peerID string, req *PushSEArtifactsRequest) error

// ReplicatorProtocol is the protocol implementation for sending resource updates to a peer node.
type ReplicatorSEProtocol struct {
	*baseProto
	processorFunc         pushSEArtifactsProcessorFunc
	replicatorFailureFunc replicatorSEFailureFunc
}

// NewReplicatorProtocol returns and a new [ReplicatorSEProtocol] struct and registers the protocol
// on the stream handler.
func NewSEReplicatorProtocol(
	h client.Host,
	pushProcessorFunc pushSEArtifactsProcessorFunc,
	replicatorFailureFunc replicatorSEFailureFunc,
) *ReplicatorSEProtocol {
	proto := &ReplicatorSEProtocol{
		baseProto:             newBaseProto(h),
		processorFunc:         pushProcessorFunc,
		replicatorFailureFunc: replicatorFailureFunc,
	}
	h.SetStreamHandler(replicatorSeProtocolRequest, proto.onRequest)
	h.SetStreamHandler(replicatorSeProtocolResponse, proto.onResponse)
	return proto
}

// PushToReplicator sends the push SE artifacts request to the provided peer node.
//
// Callers should set an appropriate context timeout.
func (proto *ReplicatorSEProtocol) PushToReplicator(
	ctx context.Context,
	req PushSEArtifactsRequest,
	pid string,
	isRetry bool,
) (reply *PushSEArtifactsReply, err error) {
	if proto.replicatorFailureFunc != nil {
		defer func() {
			// When the event is a retry, we don't need to republish the failure as
			// it is already being handled by the retry mechanism through the success channel.
			if err != nil && !isRetry {
				handleRepErr := proto.replicatorFailureFunc(ctx, pid, &req)
				if handleRepErr != nil {
					err = errors.Join(err, handleRepErr)
				}
			}
		}()
	}

	return message.Send[*PushSEArtifactsReply](ctx, proto, &req, pid, replicatorSeProtocolRequest)
}

func (proto *ReplicatorSEProtocol) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	req := PushSEArtifactsRequest{}
	err := message.Receive(stream, peerID, proto, &req)
	if err != nil {
		return
	}

	defer func() {
		// if an error occurs, try to tell the node that sent the request what went wrong.
		if err != nil {
			resp := PushSEArtifactsReply{}
			resp.SetMessageID(req.MessageID)
			resp.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, proto, &resp, peerID, replicatorProtocolResponse)
		}
	}()

	err = proto.processorFunc(ctx, &req, true)
	if err != nil {
		return
	}

	resp := PushSEArtifactsReply{}
	resp.SetMessageID(req.MessageID)
	err = message.SendAndForget(ctx, proto, &resp, peerID, replicatorSeProtocolResponse)
}

func (proto *ReplicatorSEProtocol) onResponse(stream io.Reader, peerID string) {
	_ = message.Receive(stream, peerID, proto, &PushSEArtifactsReply{})
}
