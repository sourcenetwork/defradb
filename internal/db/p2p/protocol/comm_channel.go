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
	"io"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

const (
	protocolVersion        = "0.0.1"
	protocolBase           = "/defradb/"
	protocolRequestSuffix  = "_req/" + protocolVersion
	protocolResponseSuffix = "_resp/" + protocolVersion
)

// CommProcessor defines the interface for processing requests
// All processor functions must return (reply, error) - this unifies the signatures
type CommProcessor[Req any, Reply any, ReqP interface {
	*Req
	message.Message
}, ReplyP interface {
	*Reply
	message.Message
}] interface {
	ProcessRequest(ctx context.Context, req Req, isReplicator bool) (Reply, error)
}

// CommChannel is the unified communication channel that replaces all three protocols
// It establishes direct communication between peers without event dependencies
//type CommChannel[Req message.Message, Reply message.Message] struct {

// T any, PT interface{ *T; message.Message }
type CommChannel[Req any, Reply any, ReqP interface {
	*Req
	message.Message
}, ReplyP interface {
	*Reply
	message.Message
}] struct {
	*baseProto
	processor        CommProcessor[Req, Reply, ReqP, ReplyP]
	requestEndpoint  string
	responseEndpoint string
}

// NewCommChannel creates a new communication channel
// This replaces NewReplicatorProtocol, NewSEReplicatorProtocol, and NewSEQueryProtocol
func NewCommChannel[Req any, Reply any, ReqP interface {
	*Req
	message.Message
}, ReplyP interface {
	*Reply
	message.Message
}](
	h client.Host,
	name string,
	processor CommProcessor[Req, Reply, ReqP, ReplyP],
) *CommChannel[Req, Reply, ReqP, ReplyP] {
	channel := &CommChannel[Req, Reply, ReqP, ReplyP]{
		baseProto:        newBaseProto(h),
		processor:        processor,
		requestEndpoint:  protocolBase + name + protocolRequestSuffix,
		responseEndpoint: protocolBase + name + protocolResponseSuffix,
	}

	h.SetStreamHandler(channel.requestEndpoint, channel.onRequest)
	h.SetStreamHandler(channel.responseEndpoint, channel.onResponse)

	return channel
}

// SendRequest sends any request to a peer and returns the reply
// This replaces all PushToReplicator methods and removes event.Update dependency
func (c *CommChannel[Req, Reply, ReqP, ReplyP]) SendRequest(
	ctx context.Context,
	req Req,
	peerID string,
	isRetry bool,
) (Reply, error) {
	reqPtr := ReqP(&req) // Convert to pointer type for message interface
	replyPtr, err := message.Send[ReplyP](ctx, c, reqPtr, peerID, c.requestEndpoint)
	if err != nil {
		var nilReply Reply
		return nilReply, err
	}

	// Dereference the pointer to get the value type
	reply := *replyPtr

	return reply, nil
}

func (c *CommChannel[Req, Reply, ReqP, ReplyP]) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()

	// Create stack-allocated request and convert to pointer that satisfies interface
	var req Req
	reqPtr := ReqP(&req) // Convert to pointer type that implements message.Message
	err := message.Receive(stream, peerID, c, reqPtr)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			var resp Reply
			respPtr := ReplyP(&resp) // Convert to pointer type that implements message.Message
			respPtr.SetMessageID(reqPtr.GetMessageID())
			respPtr.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, c, respPtr, peerID, c.responseEndpoint)
		}
	}()

	reply, err := c.processor.ProcessRequest(ctx, req, true)
	if err != nil {
		return
	}

	// Set message ID and send response
	replyPtr := ReplyP(&reply) // Convert to pointer type for message interface
	replyPtr.SetMessageID(reqPtr.GetMessageID())
	err = message.SendAndForget(ctx, c, replyPtr, peerID, c.responseEndpoint)
}

func (c *CommChannel[Req, Reply, ReqP, ReplyP]) onResponse(stream io.Reader, peerID string) {
	var reply Reply
	replyPtr := ReplyP(&reply) // Convert to pointer type for message interface
	_ = message.Receive(stream, peerID, c, replyPtr)
}
