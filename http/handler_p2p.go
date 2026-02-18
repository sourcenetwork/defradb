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
	"context"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

type p2pHandler struct{}

func (h *p2pHandler) PeerInfo(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	opt := options.WithIdentity(options.PeerInfo(), identity.FromContext(req.Context()))
	addresses, err := db.PeerInfo(req.Context(), opt)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, addresses)
}

func (h *p2pHandler) ActivePeers(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	opt := options.WithIdentity(options.ActivePeers(), identity.FromContext(req.Context()))
	peers, err := db.ActivePeers(req.Context(), opt)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, peers)
}

func (h *p2pHandler) Connect(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var resp []string
	if err := requestJSON(req, &resp); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.Connect(), identity.FromContext(ctx))

	err := db.Connect(ctx, resp, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) AddReplicator(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var rep AddReplicatorParams
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(
		options.AddReplicator().SetCollectionNames(rep.Collections),
		identity.FromContext(ctx),
	)

	err := db.AddReplicator(ctx, rep.Addresses, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) DeleteReplicator(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var rep DeleteReplicatorParams
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(
		options.DeleteReplicator().SetCollectionNames(rep.Collections),
		identity.FromContext(ctx),
	)

	err := db.DeleteReplicator(ctx, rep.ID, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) ListReplicators(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	opt := options.WithIdentity(options.ListReplicators(), identity.FromContext(ctx))

	reps, err := db.ListReplicators(ctx, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, reps)
}

func (h *p2pHandler) AddP2PCollections(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.AddP2PCollections(), identity.FromContext(ctx))

	err := db.AddP2PCollections(ctx, collectionIDs, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) DeleteP2PCollections(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var collectionIDs []string
	if err := requestJSON(req, &collectionIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.DeleteP2PCollections(), identity.FromContext(ctx))

	err := db.DeleteP2PCollections(ctx, collectionIDs, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) ListP2PCollections(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	opt := options.WithIdentity(options.ListP2PCollections(), identity.FromContext(ctx))

	cols, err := db.ListP2PCollections(ctx, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (h *p2pHandler) AddP2PDocuments(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var docIDs []string
	if err := requestJSON(req, &docIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.AddP2PDocuments(), identity.FromContext(ctx))

	err := db.AddP2PDocuments(ctx, docIDs, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) DeleteP2PDocuments(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	var docIDs []string
	if err := requestJSON(req, &docIDs); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opt := options.WithIdentity(options.DeleteP2PDocuments(), identity.FromContext(ctx))

	err := db.DeleteP2PDocuments(ctx, docIDs, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) ListP2PDocuments(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()

	opt := options.WithIdentity(options.ListP2PDocuments(), identity.FromContext(ctx))

	docIDs, err := db.ListP2PDocuments(ctx, opt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docIDs)
}

func (h *p2pHandler) SyncDocuments(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var reqBody struct {
		CollectionName string   `json:"collectionName"`
		DocIDs         []string `json:"docIDs"`
		Timeout        string   `json:"timeout"`
	}

	if err := requestJSON(req, &reqBody); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	ctx := req.Context()
	if reqBody.Timeout != "" {
		timeout, err := time.ParseDuration(reqBody.Timeout)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	err := db.SyncDocuments(ctx, reqBody.CollectionName, reqBody.DocIDs)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) SyncCollectionVersions(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var reqBody struct {
		VersionIDs []string `json:"versionIDs"`
		Timeout    string   `json:"timeout"`
	}

	if err := requestJSON(req, &reqBody); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	ctx := req.Context()
	if reqBody.Timeout != "" {
		timeout, err := time.ParseDuration(reqBody.Timeout)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	opts := options.WithIdentity(options.SyncCollectionVersions(), identity.FromContext(ctx))
	err := db.SyncCollectionVersions(ctx, reqBody.VersionIDs, opts)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *p2pHandler) SyncBranchableCollection(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)

	var reqBody struct {
		CollectionID string `json:"collectionID"`
		Timeout      string `json:"timeout"`
	}

	if err := requestJSON(req, &reqBody); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	ctx := req.Context()
	if reqBody.Timeout != "" {
		timeout, err := time.ParseDuration(reqBody.Timeout)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	opts := options.WithIdentity(options.SyncBranchableCollection(), identity.FromContext(ctx))
	err := db.SyncBranchableCollection(ctx, reqBody.CollectionID, opts)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
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
	addReplicatorParamsSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/add_replicator_params",
	}
	deleteReplicatorParamsSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/delete_replicator_params",
	}

	peerInfoResponse := openapi3.NewResponse().
		WithDescription("Peer network info").
		WithContent(openapi3.NewContentWithJSONSchemaRef(peerInfoSchema))

	peerInfo := openapi3.NewOperation()
	peerInfo.OperationID = "peer_info"
	peerInfo.Tags = []string{"p2p"}
	peerInfo.AddResponse(200, peerInfoResponse)
	peerInfo.Responses.Set("400", errorResponse)

	activePeersSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	activePeersResponse := openapi3.NewResponse().
		WithDescription("Connected peers").
		WithContent(openapi3.NewContentWithJSONSchema(activePeersSchema))

	activePeers := openapi3.NewOperation()
	activePeers.OperationID = "active_peers"
	activePeers.Tags = []string{"p2p"}
	activePeers.AddResponse(200, activePeersResponse)
	activePeers.Responses.Set("400", errorResponse)

	connect := openapi3.NewOperation()
	connect.OperationID = "connect"
	connect.Tags = []string{"p2p"}
	connect.Responses = openapi3.NewResponses()
	connect.Responses.Set("200", successResponse)
	connect.Responses.Set("400", errorResponse)

	listReplicatorsSchema := openapi3.NewArraySchema()
	listReplicatorsSchema.Items = replicatorSchema
	listReplicatorsResponse := openapi3.NewResponse().
		WithDescription("Replicators").
		WithContent(openapi3.NewContentWithJSONSchema(listReplicatorsSchema))

	listReplicators := openapi3.NewOperation()
	listReplicators.Description = "List peer replicators"
	listReplicators.OperationID = "peer_replicator_list"
	listReplicators.Tags = []string{"p2p"}
	listReplicators.AddResponse(200, listReplicatorsResponse)
	listReplicators.Responses.Set("400", errorResponse)

	addReplicatorRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(addReplicatorParamsSchema))

	addReplicator := openapi3.NewOperation()
	addReplicator.Description = "Add peer replicators"
	addReplicator.OperationID = "peer_replicator_add"
	addReplicator.Tags = []string{"p2p"}
	addReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: addReplicatorRequest,
	}
	addReplicator.Responses = openapi3.NewResponses()
	addReplicator.Responses.Set("200", successResponse)
	addReplicator.Responses.Set("400", errorResponse)

	deleteReplicatorRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(deleteReplicatorParamsSchema))

	deleteReplicator := openapi3.NewOperation()
	deleteReplicator.Description = "Delete peer replicators"
	deleteReplicator.OperationID = "peer_replicator_delete"
	deleteReplicator.Tags = []string{"p2p"}
	deleteReplicator.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteReplicatorRequest,
	}
	deleteReplicator.Responses = openapi3.NewResponses()
	deleteReplicator.Responses.Set("200", successResponse)
	deleteReplicator.Responses.Set("400", errorResponse)

	peerCollectionsSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	peerCollectionRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	listPeerCollectionsResponse := openapi3.NewResponse().
		WithDescription("Peer collections").
		WithContent(openapi3.NewContentWithJSONSchema(peerCollectionsSchema))

	listPeerCollections := openapi3.NewOperation()
	listPeerCollections.Description = "List peer collections"
	listPeerCollections.OperationID = "peer_collections_list"
	listPeerCollections.Tags = []string{"p2p"}
	listPeerCollections.AddResponse(200, listPeerCollectionsResponse)
	listPeerCollections.Responses.Set("400", errorResponse)

	addPeerCollections := openapi3.NewOperation()
	addPeerCollections.Description = "Add peer collections"
	addPeerCollections.OperationID = "peer_collections_add"
	addPeerCollections.Tags = []string{"p2p"}
	addPeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	addPeerCollections.Responses = openapi3.NewResponses()
	addPeerCollections.Responses.Set("200", successResponse)
	addPeerCollections.Responses.Set("400", errorResponse)

	deletePeerCollections := openapi3.NewOperation()
	deletePeerCollections.Description = "Delete peer collections"
	deletePeerCollections.OperationID = "peer_collections_delete"
	deletePeerCollections.Tags = []string{"p2p"}
	deletePeerCollections.RequestBody = &openapi3.RequestBodyRef{
		Value: peerCollectionRequest,
	}
	deletePeerCollections.Responses = openapi3.NewResponses()
	deletePeerCollections.Responses.Set("200", successResponse)
	deletePeerCollections.Responses.Set("400", errorResponse)

	peerDocumentsSchema := openapi3.NewArraySchema().
		WithItems(openapi3.NewStringSchema())

	peerDocumentRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(peerDocumentsSchema))

	listPeerDocumentsResponse := openapi3.NewResponse().
		WithDescription("Peer documents").
		WithContent(openapi3.NewContentWithJSONSchema(peerDocumentsSchema))

	listPeerDocuments := openapi3.NewOperation()
	listPeerDocuments.Description = "List peer documents"
	listPeerDocuments.OperationID = "peer_documents_list"
	listPeerDocuments.Tags = []string{"p2p"}
	listPeerDocuments.AddResponse(200, listPeerDocumentsResponse)
	listPeerDocuments.Responses.Set("400", errorResponse)

	addPeerDocuments := openapi3.NewOperation()
	addPeerDocuments.Description = "Add peer documents"
	addPeerDocuments.OperationID = "peer_documents_add"
	addPeerDocuments.Tags = []string{"p2p"}
	addPeerDocuments.RequestBody = &openapi3.RequestBodyRef{
		Value: peerDocumentRequest,
	}
	addPeerDocuments.Responses = openapi3.NewResponses()
	addPeerDocuments.Responses.Set("200", successResponse)
	addPeerDocuments.Responses.Set("400", errorResponse)

	deletePeerDocuments := openapi3.NewOperation()
	deletePeerDocuments.Description = "Delete peer documents"
	deletePeerDocuments.OperationID = "peer_documents_delete"
	deletePeerDocuments.Tags = []string{"p2p"}
	deletePeerDocuments.RequestBody = &openapi3.RequestBodyRef{
		Value: peerDocumentRequest,
	}
	deletePeerDocuments.Responses = openapi3.NewResponses()
	deletePeerDocuments.Responses.Set("200", successResponse)
	deletePeerDocuments.Responses.Set("400", errorResponse)

	syncDocumentsRequestSchema := openapi3.NewObjectSchema().
		WithProperty("collectionName", openapi3.NewStringSchema()).
		WithProperty("docIDs", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
		WithProperty("timeout", openapi3.NewStringSchema())

	syncDocumentsRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(syncDocumentsRequestSchema))

	syncDocumentsResponse := openapi3.NewResponse().
		WithDescription("Document sync completed successfully")

	syncDocuments := openapi3.NewOperation()
	syncDocuments.Description = "Synchronize documents from the network"
	syncDocuments.OperationID = "peer_sync_documents"
	syncDocuments.Tags = []string{"p2p"}
	syncDocuments.RequestBody = &openapi3.RequestBodyRef{
		Value: syncDocumentsRequest,
	}
	syncDocuments.Responses = openapi3.NewResponses()
	syncDocuments.Responses.Set("200", &openapi3.ResponseRef{Value: syncDocumentsResponse})
	syncDocuments.Responses.Set("400", errorResponse)
	syncDocuments.Responses.Set("500", errorResponse)

	syncCollectionVersionsRequestSchema := openapi3.NewObjectSchema().
		WithProperty("versionIDs", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
		WithProperty("timeout", openapi3.NewStringSchema())

	syncCollectionVersionsRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(syncCollectionVersionsRequestSchema))

	syncCollectionVersionsResponse := openapi3.NewResponse().
		WithDescription("Collection synchronization completed successfully")

	syncCollectionVersions := openapi3.NewOperation()
	syncCollectionVersions.Description = "Synchronize collection versions to the local node"
	syncCollectionVersions.OperationID = "peer_sync_collection_versions"
	syncCollectionVersions.Tags = []string{"p2p"}
	syncCollectionVersions.RequestBody = &openapi3.RequestBodyRef{
		Value: syncCollectionVersionsRequest,
	}
	syncCollectionVersions.Responses = openapi3.NewResponses()
	syncCollectionVersions.Responses.Set("200", &openapi3.ResponseRef{Value: syncCollectionVersionsResponse})
	syncCollectionVersions.Responses.Set("400", errorResponse)
	syncCollectionVersions.Responses.Set("500", errorResponse)

	syncBranchableCollectionRequestSchema := openapi3.NewObjectSchema().
		WithProperty("collectionName", openapi3.NewStringSchema()).
		WithProperty("timeout", openapi3.NewStringSchema())

	syncBranchableCollectionRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(syncBranchableCollectionRequestSchema))

	syncBranchableCollectionResponse := openapi3.NewResponse().
		WithDescription("Branchable collection sync completed successfully")

	syncBranchableCollection := openapi3.NewOperation()
	syncBranchableCollection.Description = "Synchronize a branchable collection's DAG from the network"
	syncBranchableCollection.OperationID = "peer_sync_branchable_collection"
	syncBranchableCollection.Tags = []string{"p2p"}
	syncBranchableCollection.RequestBody = &openapi3.RequestBodyRef{
		Value: syncBranchableCollectionRequest,
	}
	syncBranchableCollection.Responses = openapi3.NewResponses()
	syncBranchableCollection.Responses.Set("200", &openapi3.ResponseRef{Value: syncBranchableCollectionResponse})
	syncBranchableCollection.Responses.Set("400", errorResponse)
	syncBranchableCollection.Responses.Set("500", errorResponse)

	router.AddRoute("/p2p/info", http.MethodGet, peerInfo, h.PeerInfo)
	router.AddRoute("/p2p/active-peers", http.MethodGet, activePeers, h.ActivePeers)
	router.AddRoute("/p2p/connect", http.MethodPost, connect, h.Connect)
	router.AddRoute("/p2p/replicators", http.MethodGet, listReplicators, h.ListReplicators)
	router.AddRoute("/p2p/replicators", http.MethodPost, addReplicator, h.AddReplicator)
	router.AddRoute("/p2p/replicators", http.MethodDelete, deleteReplicator, h.DeleteReplicator)
	router.AddRoute("/p2p/collections", http.MethodGet, listPeerCollections, h.ListP2PCollections)
	router.AddRoute("/p2p/collections", http.MethodPost, addPeerCollections, h.AddP2PCollections)
	router.AddRoute("/p2p/collections", http.MethodDelete, deletePeerCollections, h.DeleteP2PCollections)
	router.AddRoute("/p2p/collections/sync-versions", http.MethodPost, syncCollectionVersions, h.SyncCollectionVersions)
	router.AddRoute("/p2p/collections/sync-branchable", http.MethodPost, syncBranchableCollection,
		h.SyncBranchableCollection)
	router.AddRoute("/p2p/documents", http.MethodGet, listPeerDocuments, h.ListP2PDocuments)
	router.AddRoute("/p2p/documents", http.MethodPost, addPeerDocuments, h.AddP2PDocuments)
	router.AddRoute("/p2p/documents", http.MethodDelete, deletePeerDocuments, h.DeleteP2PDocuments)
	router.AddRoute("/p2p/documents/sync", http.MethodPost, syncDocuments, h.SyncDocuments)
}
