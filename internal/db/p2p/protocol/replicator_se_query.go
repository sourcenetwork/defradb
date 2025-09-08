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

/*
// pushSEArtifacts creates and sends SE artifacts to another node
func (s *server) pushSEArtifacts(evt se.ReplicateEvent, pid peer.ID) (err error) {
	defer func() {
		if err != nil && !evt.IsRetry {
			// Collect unique field names from artifacts
			fieldNamesMap := make(map[string]struct{})
			for _, artifact := range evt.Artifacts {
				fieldNamesMap[artifact.FieldName] = struct{}{}
			}

			var fieldNames []string
			for fieldName := range fieldNamesMap {
				fieldNames = append(fieldNames, fieldName)
			}

			s.peer.bus.Publish(event.NewMessage(se.ReplicationFailureEventName, se.ReplicationFailureEvent{
				DocID:        evt.DocID,
				CollectionID: evt.CollectionID,
				PeerID:       pid,
				FieldNames:   fieldNames,
				Identity:     evt.Identity,
			}))
		}
		if evt.Success != nil {
			evt.Success <- err == nil
		}
	}()

	client, err := s.dial(pid)
	if err != nil {
		return NewErrPushSEArtifacts(err)
	}

	ctx, cancel := context.WithTimeout(s.peer.ctx, PushTimeout)
	defer cancel()

	netArtifacts := make([]seArtifact, len(evt.Artifacts))
	for i, artifact := range evt.Artifacts {
		netArtifacts[i] = seArtifact{
			DocID:     artifact.DocID,
			IndexID:   artifact.IndexID,
			SearchTag: artifact.SearchTag,
		}
	}

	log.InfoContext(ctx, "Handle push SE artifacts",
		corelog.String("DocID", evt.DocID),
		corelog.String("CollectionID", evt.CollectionID),
		corelog.String("PeerID", pid.String()))

	req := pushSEArtifactsRequest{
		CollectionID: evt.CollectionID,
		Artifacts:    netArtifacts,
		Creator:      s.peer.host.ID().String(),
	}

	if err := client.Invoke(ctx, servicePushSEArtifactsName, req, nil); err != nil {
		return NewErrPushSEArtifacts(
			err,
			errors.NewKV("DocID", evt.DocID),
			errors.NewKV("CollectionID", evt.CollectionID),
			errors.NewKV("PeerID", pid),
		)
	}
	return nil
}

// querySEArtifacts queries SE artifacts on a remote node
func (s *server) querySEArtifacts(ctx context.Context, pid peer.ID, req querySEArtifactsRequest) (*querySEArtifactsReply, error) {
	client, err := s.dial(pid)
	if err != nil {
		return nil, NewErrQuerySEArtifacts(err)
	}

	ctx, cancel := context.WithTimeout(ctx, PullTimeout)
	defer cancel()

	resp := &querySEArtifactsReply{}
	if err := client.Invoke(ctx, serviceQuerySEArtifactsName, req, resp); err != nil {
		return nil, NewErrQuerySEArtifacts(err,
			errors.NewKV("CollectionID", req.CollectionID),
			errors.NewKV("PeerID", pid),
		)
	}

	return resp, nil
}
*/

/*
// pushSEArtifactsHandler receives SE artifacts from peers
func (s *server) pushSEArtifactsHandler(ctx context.Context, req *pushSEArtifactsRequest) (*pushSEArtifactsReply, error) {
	sb := strings.Builder{}
	for i, netArtifact := range req.Artifacts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(netArtifact.DocID)
	}
	log.InfoContext(ctx, "Handle push SE artifacts", corelog.String("DocIDs", sb.String()), corelog.String("Sender", req.Creator))

	_, err := peerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	artifacts := make([]secore.Artifact, len(req.Artifacts))
	for i, netArtifact := range req.Artifacts {
		artifacts[i] = secore.Artifact{
			DocID:        netArtifact.DocID,
			IndexID:      netArtifact.IndexID,
			SearchTag:    netArtifact.SearchTag,
			CollectionID: req.CollectionID,
		}
	}

	// Store artifacts directly in the datastore
	if err := se.StoreArtifacts(ctx, datastore.DatastoreFrom(s.peer.db.Rootstore()), artifacts); err != nil {
		return nil, err
	}

	return &pushSEArtifactsReply{}, nil
}

// querySEArtifactsHandler handles SE queries from peers
func (s *server) querySEArtifactsHandler(ctx context.Context, req *querySEArtifactsRequest) (*querySEArtifactsReply, error) {
	peerID, err := peerIDFromContext(ctx)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to get peer ID from context", err)
		return nil, err
	}

	matchingDocIDs, err := s.querySEArtifactsFromDatastore(ctx, req)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to query SE artifacts from datastore", err)
		return nil, err
	}

	log.InfoContext(ctx, "Handle SE artifacts query", corelog.String("DocIDs", strings.Join(matchingDocIDs, ", ")),
		corelog.String("Sender", peerID.String()))

	return &querySEArtifactsReply{
		DocIDs: matchingDocIDs,
	}, nil
}

// querySEArtifactsFromDatastore queries SE artifacts from the local datastore
func (s *server) querySEArtifactsFromDatastore(ctx context.Context, req *querySEArtifactsRequest) ([]string, error) {
	queries := make([]se.FieldQuery, len(req.Queries))
	for i, q := range req.Queries {
		queries[i] = se.FieldQuery{
			FieldName: q.FieldName,
			IndexID:   q.IndexID,
			SearchTag: q.SearchTag,
		}
	}

	return se.FetchDocIDs(ctx, datastore.DatastoreFrom(s.peer.db.Rootstore()), req.CollectionID, queries)
}
*/

/* // from grpc.go

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

*/
