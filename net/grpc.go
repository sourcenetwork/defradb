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

	servicePushLogName           = "/" + grpcServiceName + "/PushLog"
	serviceGetIdentityName       = "/" + grpcServiceName + "/GetIdentity"
	servicePushSEArtifactsName   = "/" + grpcServiceName + "/PushSEArtifacts"
	serviceQuerySEArtifactsName  = "/" + grpcServiceName + "/QuerySEArtifacts"
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

// pushSEArtifactsRequest - Request to push SE artifacts
type pushSEArtifactsRequest struct {
	CollectionID string
	Artifacts    []seArtifact
	Creator      string
}

// seArtifact - Network representation
type seArtifact struct {
	DocID     string
	IndexID   string
	SearchTag []byte
}

// Reply type
type pushSEArtifactsReply struct{}

// querySEArtifactsRequest - Request to query SE artifacts
type querySEArtifactsRequest struct {
	CollectionID string
	Queries      []seFieldQuery
}

// seFieldQuery - Query for a specific field
type seFieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// querySEArtifactsReply - Reply with matching document IDs
type querySEArtifactsReply struct {
	DocIDs []string
}

type serviceServer interface {
	// pushLogHandler handles a push log request to sync blocks.
	pushLogHandler(context.Context, *pushLogRequest) (*pushLogReply, error)
	// getIdentityHandler handles an indentity request and returns the local node's identity.
	getIdentityHandler(context.Context, *getIdentityRequest) (*getIdentityReply, error)
	// pushSEArtifactsHandler handles SE artifacts push request.
	pushSEArtifactsHandler(context.Context, *pushSEArtifactsRequest) (*pushSEArtifactsReply, error)
	// querySEArtifactsHandler handles SE artifacts query request.
	querySEArtifactsHandler(context.Context, *querySEArtifactsRequest) (*querySEArtifactsReply, error)
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

func pushSEArtifactsHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	in := new(pushSEArtifactsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(serviceServer).pushSEArtifactsHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: servicePushSEArtifactsName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(serviceServer).pushSEArtifactsHandler(ctx, req.(*pushSEArtifactsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func querySEArtifactsHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	in := new(querySEArtifactsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(serviceServer).querySEArtifactsHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: serviceQuerySEArtifactsName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(serviceServer).querySEArtifactsHandler(ctx, req.(*querySEArtifactsRequest))
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
			{
				MethodName: "PushSEArtifacts",
				Handler:    pushSEArtifactsHandler,
			},
			{
				MethodName: "QuerySEArtifacts",
				Handler:    querySEArtifactsHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "defradb.cbor",
	}
	s.RegisterService(desc, srv)
}
