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

const (
	replicatorSeQueryProtocolVersion  = "0.0.1"
	replicatorSeQueryProtocolRequest  = "/defradb/rep_se_query_req/" + replicatorSeQueryProtocolVersion
	replicatorSeQueryProtocolResponse = "/defradb/rep_se_query_resp/" + replicatorSeQueryProtocolVersion
)

type QuerySEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Queries      []SEFieldQuery
}

// seFieldQuery - Query for a specific field
type SEFieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// QuerySEArtifactsReply - Reply with matching document IDs
type QuerySEArtifactsReply struct {
	message.MetaData
	DocIDs []string
}

type querySEArtifactsProcessorFunc func(
	ctx context.Context,
	req *QuerySEArtifactsRequest,
	isReplicator bool,
) (QuerySEArtifactsReply, error)

type replicatorSEQueryFailureFunc func(ctx context.Context, peerID string, req *QuerySEArtifactsRequest) error

// ReplicatorProtocol is the protocol implementation for sending resource updates to a peer node.
type SEQueryProtocol struct {
	*baseProto
	processorFunc         querySEArtifactsProcessorFunc
	replicatorFailureFunc replicatorSEQueryFailureFunc
}

// NewReplicatorProtocol returns and a new [ReplicatorSEProtocol] struct and registers the protocol
// on the stream handler.
func NewSEQueryProtocol(
	h client.Host,
	pushProcessorFunc querySEArtifactsProcessorFunc,
	replicatorFailureFunc replicatorSEQueryFailureFunc,
) *SEQueryProtocol {
	proto := &SEQueryProtocol{
		baseProto:             newBaseProto(h),
		processorFunc:         pushProcessorFunc,
		replicatorFailureFunc: replicatorFailureFunc,
	}
	h.SetStreamHandler(replicatorSeQueryProtocolRequest, proto.onRequest)
	h.SetStreamHandler(replicatorSeQueryProtocolResponse, proto.onResponse)
	return proto
}

// PushToReplicator sends the pushseartifacts request to the provided peer node.
//
// Callers should set an appropriate context timeout.
func (proto *SEQueryProtocol) PushToReplicator(
	ctx context.Context,
	req QuerySEArtifactsRequest,
	pid string,
	isRetry bool,
) (reply *QuerySEArtifactsReply, err error) {
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

	return message.Send[*QuerySEArtifactsReply](ctx, proto, &req, pid, replicatorProtocolRequest)
}

func (proto *SEQueryProtocol) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	req := QuerySEArtifactsRequest{}
	err := message.Receive(stream, peerID, proto, &req)
	if err != nil {
		return
	}

	defer func() {
		// if an error occurs, try to tell the node that sent the request what went wrong.
		if err != nil {
			resp := QuerySEArtifactsReply{}
			resp.SetMessageID(req.MessageID)
			resp.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, proto, &resp, peerID, replicatorProtocolResponse)
		}
	}()

	resp, err := proto.processorFunc(ctx, &req, true)
	if err != nil {
		return
	}

	resp.SetMessageID(req.MessageID)
	err = message.SendAndForget(ctx, proto, &resp, peerID, replicatorProtocolResponse)
}

func (proto *SEQueryProtocol) onResponse(stream io.Reader, peerID string) {
	_ = message.Receive(stream, peerID, proto, &QuerySEArtifactsReply{})
}

