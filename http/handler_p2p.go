// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/defradb/client"
)

type p2pHandler struct{}

func (s *p2pHandler) PeerInfo(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	responseJSON(rw, http.StatusOK, p2p.PeerInfo())
}

func (s *p2pHandler) SetReplicator(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	var rep client.ReplicatorParams
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.SetReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) DeleteReplicator(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	var rep client.ReplicatorParams
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.DeleteReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) GetAllReplicators(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	reps, err := p2p.GetAllReplicators(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, reps)
}

func (s *p2pHandler) AddP2PCollection(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.AddP2PCollections(req.Context(), collectionIDs)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) RemoveP2PCollection(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.RemoveP2PCollections(req.Context(), collectionIDs)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) GetAllP2PCollections(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := tryGetContextClientP2P(req)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}

	cols, err := p2p.GetAllP2PCollections(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (h *p2pHandler) bindRoutes(router *Router) {
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	peerInfoSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/peer_info",
	}
	replicatorSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/replicator",
	}
	replicatorParamsSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/replicator_params",
	}

	peerInfoResponse := openapi3.NewResponse().
		WithDescription("Peer network info").
		WithContent(openapi3.NewContentWithJSONSchemaRef(peerInfoSchema))

	peerInfo := openapi3.NewOperation()
	peerInfo.OperationID = "peer_info"
	peerInfo.Tags = []string{"p2p"}
	peerInfo.AddResponse(200, peerInfoResponse)
	peerInfo.Responses.Set("400", errorResponse)

	getReplicatorsSchema := openapi3.NewArraySchema()
	getReplicatorsSchema.Items = replicatorSchema
	getReplicatorsResponse := openapi3.NewResponse().
		WithDescription("Replicators").
		WithContent(openapi3.NewContentWithJSONSchema(getReplicatorsSchema))

	getReplicators := openapi3.NewOperation()
	getReplicators.Description = "List peer replicators"
	getReplicators.OperationID = "peer_replicator_list"
	getReplicators.Tags = []string{"p2p"}
	getReplicators.AddResponse(200, getReplicatorsResponse)
	getReplicators.Responses.Set("400", errorResponse)

	replicatorRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(replicatorParamsSchema))

	setReplicator := openapi3.NewOperation()
	setReplicator.Description = "Add peer replicators"
	setReplicator.OperationID = "peer_replicator_set"
	setReplicator.Tags = []string{"p2p"}
	setReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	setReplicator.Responses = openapi3.NewResponses()
	setReplicator.Responses.Set("200", successResponse)
	setReplicator.Responses.Set("400", errorResponse)

	deleteReplicator := openapi3.NewOperation()
	deleteReplicator.Description = "Delete peer replicators"
	deleteReplicator.OperationID = "peer_replicator_delete"
	deleteReplicator.Tags = []string{"p2p"}
	deleteReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: replicatorRequest,
	}
	deleteReplicator.Responses = openapi3.NewResponses()
	deleteReplicator.Responses.Set("200", successResponse)
	deleteReplicator.Responses.Set("400", errorResponse)

	peerCollectionsSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	peerCollectionRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollectionsResponse := openapi3.NewResponse().
		WithDescription("Peer collections").
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	getPeerCollections := openapi3.NewOperation()
	getPeerCollections.Description = "List peer collections"
	getPeerCollections.OperationID = "peer_collection_list"
	getPeerCollections.Tags = []string{"p2p"}
	getPeerCollections.AddResponse(200, getPeerCollectionsResponse)
	getPeerCollections.Responses.Set("400", errorResponse)

	addPeerCollections := openapi3.NewOperation()
	addPeerCollections.Description = "Add peer collections"
	addPeerCollections.OperationID = "peer_collection_add"
	addPeerCollections.Tags = []string{"p2p"}
	addPeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	addPeerCollections.Responses = openapi3.NewResponses()
	addPeerCollections.Responses.Set("200", successResponse)
	addPeerCollections.Responses.Set("400", errorResponse)

	removePeerCollections := openapi3.NewOperation()
	removePeerCollections.Description = "Remove peer collections"
	removePeerCollections.OperationID = "peer_collection_remove"
	removePeerCollections.Tags = []string{"p2p"}
	removePeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	removePeerCollections.Responses = openapi3.NewResponses()
	removePeerCollections.Responses.Set("200", successResponse)
	removePeerCollections.Responses.Set("400", errorResponse)

	router.AddRoute("/p2p/info", http.MethodGet, peerInfo, h.PeerInfo)
	router.AddRoute("/p2p/replicators", http.MethodGet, getReplicators, h.GetAllReplicators)
	router.AddRoute("/p2p/replicators", http.MethodPost, setReplicator, h.SetReplicator)
	router.AddRoute("/p2p/replicators", http.MethodDelete, deleteReplicator, h.DeleteReplicator)
	router.AddRoute("/p2p/collections", http.MethodGet, getPeerCollections, h.GetAllP2PCollections)
	router.AddRoute("/p2p/collections", http.MethodPost, addPeerCollections, h.AddP2PCollection)
	router.AddRoute("/p2p/collections", http.MethodDelete, removePeerCollections, h.RemoveP2PCollection)
}
