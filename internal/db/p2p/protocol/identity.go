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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

const (
	identityProtocolVersion  = "0.0.1"
	identityProtocolRequest  = "/defradb/ident_req/" + identityProtocolVersion
	identityProtocolResponse = "/defradb/ident_resp/" + identityProtocolVersion
)

// IdentityRequest is the struct used to request the identity of a peer node.
type IdentityRequest struct {
	message.MetaData
	// PeerID is the ID of the requesting peer.
	// It will be used as the audience for the identity token.
	PeerID string
}

// IdentityResponse is the expected response struct that should be received after
// an identity request.
type IdentityResponse struct {
	message.MetaData
	// IdentityToken is the token that can be used to authenticate the peer.
	IdentityToken []byte
}

type getIdentityFunc func(ctx context.Context, audience immutable.Option[string]) ([]byte, error)

// IdentityProtocol is the protocol implementation for requesting an identity from a peer node.
type IdentityProtocol struct {
	*baseProto
	getIdentityFunc getIdentityFunc
}

// NewIdentityProtocol returns and a new [IdentityProtocol] struct and registers the protocol
// on the stream handler.
func NewIdentityProtocol(h client.Host, getIdentityFunc getIdentityFunc) *IdentityProtocol {
	proto := &IdentityProtocol{
		baseProto:       newBaseProto(h),
		getIdentityFunc: getIdentityFunc,
	}
	h.SetStreamHandler(identityProtocolRequest, proto.onRequest)
	h.SetStreamHandler(identityProtocolResponse, proto.onResponse)
	return proto
}

// GetIdentity sends the identity request to the provided peer node.
//
// Callers should set an appropriate context timeout.
func (proto *IdentityProtocol) GetIdentity(ctx context.Context, pid string) (*IdentityResponse, error) {
	req := IdentityRequest{
		PeerID: proto.host.ID(),
	}
	return message.Send[*IdentityResponse](ctx, proto, &req, pid, identityProtocolRequest)
}

func (proto *IdentityProtocol) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	req := IdentityRequest{}
	err := message.Receive(stream, peerID, proto, &req)
	if err != nil {
		return
	}
	defer func() {
		// if an error occurs, try to tell the node that sent the request what went wrong.
		if err != nil {
			resp := IdentityResponse{}
			resp.SetMessageID(req.MessageID)
			resp.SetErrMessage(err.Error())
			_ = message.SendAndForget(ctx, proto, &resp, peerID, identityProtocolResponse)
		}
	}()
	token, err := proto.getIdentityFunc(ctx, immutable.Some(req.PeerID))
	if err != nil {
		return
	}
	resp := IdentityResponse{IdentityToken: token}
	resp.SetMessageID(req.MessageID)
	err = message.SendAndForget(ctx, proto, &resp, peerID, identityProtocolResponse)
}

func (proto *IdentityProtocol) onResponse(stream io.Reader, peerID string) {
	_ = message.Receive(stream, peerID, proto, &IdentityResponse{})
}
