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

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

var logCommChannel = corelog.NewLogger("protocol.comm_channel")

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
	HandleFailure(ctx context.Context, peerID string, req Req) error
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

const protocolVersion = "0.0.1"
const protocolBase = "/defradb/"
const protocolRequestSuffix = "_req/" + protocolVersion
const protocolResponseSuffix = "_resp/" + protocolVersion

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
	logCommChannel.Info(">>> comm_channel.NewCommChannel: Creating communication channel",
		corelog.String("RequestEndpoint", protocolBase+name+protocolRequestSuffix),
		corelog.String("ResponseEndpoint", protocolBase+name+protocolResponseSuffix))

		channel := &CommChannel[Req, Reply, ReqP, ReplyP]{
		baseProto:        newBaseProto(h),
		processor:        processor,
		requestEndpoint:  protocolBase + name + protocolRequestSuffix,
		responseEndpoint: protocolBase + name + protocolResponseSuffix,
	}

	h.SetStreamHandler(channel.requestEndpoint, channel.onRequest)
	h.SetStreamHandler(channel.responseEndpoint, channel.onResponse)

	logCommChannel.Info(">>> comm_channel.NewCommChannel: Channel created and handlers registered")
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
	logCommChannel.Info(">>> comm_channel.SendRequest: Starting",
		corelog.Any("PeerID", peerID),
		corelog.String("RequestEndpoint", c.requestEndpoint))

	var err error
	defer func() {
		if err != nil && !isRetry {
			logCommChannel.Info(">>> comm_channel.SendRequest: Handling failure", corelog.Any("PeerID", peerID))
			if handleErr := c.processor.HandleFailure(ctx, peerID, req); handleErr != nil {
				err = errors.Join(err, handleErr)
			}
		}
	}()

	logCommChannel.Info(">>> comm_channel.SendRequest: Sending message to peer", corelog.Any("PeerID", peerID))

	reqPtr := ReqP(&req) // Convert to pointer type for message interface
	replyPtr, err := message.Send[ReplyP](ctx, c, reqPtr, peerID, c.requestEndpoint)
	if err != nil {
		logCommChannel.ErrorE(">>> comm_channel.SendRequest: Send failed", err, corelog.Any("PeerID", peerID))
		var nilReply Reply
		return nilReply, err
	}

	logCommChannel.Info(">>> comm_channel.SendRequest: Got reply", corelog.Any("PeerID", peerID))

	// Dereference the pointer to get the value type
	reply := *replyPtr
	logCommChannel.Info(">>> comm_channel.SendRequest: Successfully returning reply", corelog.Any("PeerID", peerID))

	return reply, nil
}

func (c *CommChannel[Req, Reply, ReqP, ReplyP]) onRequest(stream io.Reader, peerID string) {
	logCommChannel.Info(">>> comm_channel.onRequest: Received request", corelog.Any("PeerID", peerID))
	ctx := context.Background()

	// Create stack-allocated request and convert to pointer that satisfies interface
	var req Req
	reqPtr := ReqP(&req) // Convert to pointer type that implements message.Message
	err := message.Receive(stream, peerID, c, reqPtr)
	if err != nil {
		logCommChannel.ErrorE(">>> comm_channel.onRequest: Failed to receive message", err, corelog.Any("PeerID", peerID))
		return
	}

	logCommChannel.Info(">>> comm_channel.onRequest: Processing request", corelog.Any("PeerID", peerID))

	defer func() {
		if err != nil {
			logCommChannel.ErrorE(">>> comm_channel.onRequest: Processing failed, sending error response", err, corelog.Any("PeerID", peerID))
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

	logCommChannel.Info(">>> comm_channel.onRequest: Got reply from processor", corelog.Any("PeerID", peerID))

	// Set message ID and send response
	replyPtr := ReplyP(&reply) // Convert to pointer type for message interface
	replyPtr.SetMessageID(reqPtr.GetMessageID())
	err = message.SendAndForget(ctx, c, replyPtr, peerID, c.responseEndpoint)
	logCommChannel.Info(">>> comm_channel.onRequest: Sent reply", corelog.Any("PeerID", peerID), corelog.Any("Error", err))
}

func (c *CommChannel[Req, Reply, ReqP, ReplyP]) onResponse(stream io.Reader, peerID string) {
	logCommChannel.Info(">>> comm_channel.onResponse: Receiving response", corelog.Any("PeerID", peerID))
	var reply Reply
	replyPtr := ReplyP(&reply) // Convert to pointer type for message interface
	err := message.Receive(stream, peerID, c, replyPtr)
	if err != nil {
		logCommChannel.ErrorE(">>> comm_channel.onResponse: Failed to receive response", err, corelog.Any("PeerID", peerID))
	} else {
		logCommChannel.Info(">>> comm_channel.onResponse: Successfully received response", corelog.Any("PeerID", peerID))
	}
}
