// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"

	"google.golang.org/grpc"
)

const (
	grpcServiceName = "defradb.net.Service"

	servicePushLogName     = "/" + grpcServiceName + "/PushLog"
	serviceGetIdentityName = "/" + grpcServiceName + "/GetIdentity"
)

type pushLogRequest struct {
	DocID        string
	CID          []byte
	CollectionID string
	Creator      string
	Block        []byte
}

type pushLogReply struct{}

type getIdentityRequest struct {
	// PeerID is the ID of the requesting peer.
	// It will be used as the audience for the identity token.
	PeerID string
}

type getIdentityReply struct {
	// IdentityToken is the token that can be used to authenticate the peer.
	IdentityToken []byte
}

// docSyncRequest represents a request to synchronize specific documents.
type docSyncRequest struct {
	DocIDs []string `json:"docIDs"`
}

// docSyncReply represents the response to a document sync request.
type docSyncReply struct {
	Results []docSyncItem `json:"results"`
	Sender  string        `json:"sender"`
}

// docSyncItem represents the sync result for a single document.
type docSyncItem struct {
	DocID string   `json:"docID"`
	Heads [][]byte `json:"heads"`
}

type serviceServer interface {
	// pushLogHandler handles a push log request to sync blocks.
	pushLogHandler(context.Context, *pushLogRequest) (*pushLogReply, error)
	// getIdentityHandler handles an indentity request and returns the local node's identity.
	getIdentityHandler(context.Context, *getIdentityRequest) (*getIdentityReply, error)
}

func getIdentityHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	in := new(getIdentityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(serviceServer).getIdentityHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: serviceGetIdentityName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(serviceServer).getIdentityHandler(ctx, req.(*getIdentityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func pushLogHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	in := new(pushLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(serviceServer).pushLogHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: servicePushLogName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(serviceServer).pushLogHandler(ctx, req.(*pushLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func registerServiceServer(s grpc.ServiceRegistrar, srv serviceServer) {
	desc := &grpc.ServiceDesc{
		ServiceName: grpcServiceName,
		HandlerType: (*serviceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "PushLog",
				Handler:    pushLogHandler,
			},
			{
				MethodName: "GetIdentity",
				Handler:    getIdentityHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "defradb.cbor",
	}
	s.RegisterService(desc, srv)
}
