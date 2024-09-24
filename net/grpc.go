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

	serviceGetDocGraphName  = "/" + grpcServiceName + "/GetDocGraph"
	servicePushDocGraphName = "/" + grpcServiceName + "/PushDocGraph"
	serviceGetLogName       = "/" + grpcServiceName + "/GetLog"
	servicePushLogName      = "/" + grpcServiceName + "/PushLog"
	serviceGetHeadLogName   = "/" + grpcServiceName + "/GetHeadLog"
)

type getDocGraphRequest struct{}

type getDocGraphReply struct{}

type getHeadLogRequest struct{}

type getHeadLogReply struct{}

type getLogRequest struct{}

type getLogReply struct{}

type pushDocGraphRequest struct{}

type pushDocGraphReply struct{}

type pushLogRequest struct {
	DocID      string
	CID        []byte
	SchemaRoot string
	Creator    string
	Block      []byte
}

type pushLogReply struct{}

type serviceServer interface {
	// GetDocGraph from this peer.
	GetDocGraph(context.Context, *getDocGraphRequest) (*getDocGraphReply, error)
	// PushDocGraph to this peer.
	PushDocGraph(context.Context, *pushDocGraphRequest) (*pushDocGraphReply, error)
	// GetLog from this peer.
	GetLog(context.Context, *getLogRequest) (*getLogReply, error)
	// PushLog to this peer.
	PushLog(context.Context, *pushLogRequest) (*pushLogReply, error)
	// GetHeadLog from this peer
	GetHeadLog(context.Context, *getHeadLogRequest) (*getHeadLogReply, error)
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
		return srv.(serviceServer).PushLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: servicePushLogName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(serviceServer).PushLog(ctx, req.(*pushLogRequest))
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
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "defradb.cbor",
	}
	s.RegisterService(desc, srv)
}
