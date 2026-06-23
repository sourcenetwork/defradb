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
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

const (
	protocolVersion        = "0.0.1"
	protocolBase           = "/defradb/"
	protocolRequestSuffix  = "_req/" + protocolVersion
	protocolResponseSuffix = "_resp/" + protocolVersion
)

type messagePointer[T any] interface {
	*T
	message.Message
}

// CommProcessor defines the interface for processing requests and replies.
// Uses 4 type parameters to solve the embedded interface pointer receiver problem:
// - Req/Reply: Value types for stack allocation and clean processor signatures
// - ReqP/ReplyP: Pointer types that implement message.Message interface
type CommProcessor[Req any, Reply any, ReqP messagePointer[Req], ReplyP messagePointer[Reply]] interface {
	ProcessRequest(ctx context.Context, req Req) (Reply, error)
}

type commChannel[Req any, Reply any, ReqP messagePointer[Req], ReplyP messagePointer[Reply]] struct {
	*baseProto
	processor        CommProcessor[Req, Reply, ReqP, ReplyP]
	requestEndpoint  string
	responseEndpoint string
}

// CommChannel defines the interface for sending requests and receiving replies.
type CommChannel[Req, Rep any] interface {
	SendRequest(context.Context, Req, string) (Rep, error)
}

// NewCommChannel creates a new communication channel [commChannel]
// 4 type parameters needed because message.MetaData implements message.Message only as pointer:
// - Req: Stack-allocated value type (e.g. PushLogRequest)
// - Reply: Stack-allocated reply type (e.g. PushLogReply)
// - ReqP: Pointer type implementing message.Message (e.g. *PushLogRequest)
// - ReplyP: Pointer type implementing message.Message (e.g. *PushLogReply)
// This enables stack allocation for performance while satisfying interface constraints.
func NewCommChannel[Req any, Reply any, ReqP messagePointer[Req], ReplyP messagePointer[Reply]](
	h client.Host,
	name string,
	processor CommProcessor[Req, Reply, ReqP, ReplyP],
) CommChannel[Req, Reply] {
	channel := &commChannel[Req, Reply, ReqP, ReplyP]{
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
func (c *commChannel[Req, Reply, ReqP, ReplyP]) SendRequest(
	ctx context.Context,
	req Req,
	peerID string,
) (Reply, error) {
	reqPtr := ReqP(&req)
	replyPtr, err := message.Send[ReplyP](ctx, c, reqPtr, peerID, c.requestEndpoint)
	if err != nil {
		var nilReply Reply
		return nilReply, err
	}

	reply := *replyPtr

	return reply, nil
}

func (c *commChannel[Req, Reply, ReqP, ReplyP]) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	ctx = id.InitCollectionShortIDCache(ctx)
	ctx = id.InitFieldShortIDCache(ctx)

	var req Req
	reqPtr := ReqP(&req)
	err := message.Receive(stream, peerID, c, reqPtr)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			var resp Reply
			respPtr := ReplyP(&resp)
			respPtr.SetMessageID(reqPtr.GetMessageID())
			respPtr.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, c, respPtr, peerID, c.responseEndpoint)
		}
	}()

	reply, err := c.processor.ProcessRequest(ctx, req)
	if err != nil {
		return
	}

	replyPtr := ReplyP(&reply)
	replyPtr.SetMessageID(reqPtr.GetMessageID())
	err = message.SendAndForget(ctx, c, replyPtr, peerID, c.responseEndpoint)
}

func (c *commChannel[Req, Reply, ReqP, ReplyP]) onResponse(stream io.Reader, peerID string) {
	var reply Reply
	replyPtr := ReplyP(&reply)
	err := message.Receive(stream, peerID, c, replyPtr)
	if err != nil {
		log.ErrorE("Failed to receive response message.", err)
	}
}
