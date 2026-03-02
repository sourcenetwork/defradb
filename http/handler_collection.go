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
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

const docEncryptParam = "encrypt"
const docEncryptFieldsParam = "encryptFields"

type collectionHandler struct{}

type CollectionDeleteRequest struct {
	Filter any `json:"filter"`
}

type CollectionUpdateRequest struct {
	Filter  any    `json:"filter"`
	Updater string `json:"updater"`
}

func (h *collectionHandler) DeleteDocumentsWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteOpt := options.WithIdentity(options.DeleteDocumentsWithFilter(), identity.FromContext(ctx))

	result, err := col.DeleteDocumentsWithFilter(ctx, request.Filter, deleteOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) UpdateDocumentsWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	updateOpt := options.WithIdentity(options.UpdateDocumentsWithFilter(), identity.FromContext(ctx))

	result, err := col.UpdateDocumentsWithFilter(ctx, request.Filter, request.Updater, updateOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) AddIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	descWithoutID := client.IndexAddRequest{
		Name:   indexDesc.Name,
		Fields: indexDesc.Fields,
		Unique: indexDesc.Unique,
	}

	addIndexOpt := options.WithIdentity(options.CollectionAddIndex(), identity.FromContext(ctx))

	index, err := col.AddIndex(ctx, descWithoutID, addIndexOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (h *collectionHandler) ListIndexes(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()
	name := chi.URLParam(req, "name")
	ident := identity.FromContext(ctx)
	col, err := db.GetCollectionByName(ctx, name, options.WithIdentity(options.GetCollectionByName(), ident))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	listIndexesOpt := options.WithIdentity(options.CollectionListIndexes(), ident)

	indexes, err := col.ListIndexes(ctx, listIndexesOpt)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (h *collectionHandler) DeleteIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	deleteIndexOpt := options.WithIdentity(options.CollectionDeleteIndex(), identity.FromContext(ctx))

	err := col.DeleteIndex(ctx, chi.URLParam(req, "index"), deleteIndexOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) AddEncryptedIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	var indexDesc client.EncryptedIndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.AddEncryptedIndex(), identity.FromContext(req.Context()))

	index, err := col.AddEncryptedIndex(req.Context(), indexDesc, opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (h *collectionHandler) ListEncryptedIndexes(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	opts := options.WithIdentity(options.CollectionListEncryptedIndexes(), identity.FromContext(req.Context()))
	indexes, err := col.ListEncryptedIndexes(req.Context(), opts)
	if err != nil {
		responseJSON(rw, http.StatusInternalServerError, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (h *collectionHandler) DeleteEncryptedIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	fieldName := chi.URLParam(req, "field")
	if fieldName == "" {
		responseJSON(rw, http.StatusBadRequest, errorResponse{fmt.Errorf("field name is required")})
		return
	}

	opts := options.WithIdentity(options.DeleteEncryptedIndex(), identity.FromContext(req.Context()))

	err := col.DeleteEncryptedIndex(req.Context(), fieldName, opts)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) Truncate(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	truncateOpt := options.WithIdentity(options.CollectionTruncate(), identity.FromContext(ctx))

	err := col.Truncate(ctx, truncateOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	collectionUpdateSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_update",
	}
	updateResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/update_result",
	}
	collectionDeleteSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/collection_delete",
	}
	deleteResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/delete_result",
	}
	documentSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/document",
	}
	indexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/index",
	}
	indexAddRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/index_add",
	}
	encryptedIndexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/encrypted_index",
	}
	encryptedIndexAddRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/encrypted_index_add",
	}

	collectionNamePathParam := openapi3.NewPathParameter("name").
		WithDescription("Collection name").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentArraySchema := openapi3.NewArraySchema()
	documentArraySchema.Items = documentSchema

	documentAddBodySchema := openapi3.NewOneOfSchema()
	documentAddBodySchema.OneOf = openapi3.SchemaRefs{
		documentSchema,
		openapi3.NewSchemaRef("", documentArraySchema),
	}

	documentAddRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(documentAddBodySchema))

	documentAdd := openapi3.NewOperation()
	documentAdd.OperationID = "document_add"
	documentAdd.Description = "Add document(s) to a collection"
	documentAdd.Tags = []string{"document"}
	documentAdd.AddParameter(collectionNamePathParam)
	documentAdd.RequestBody = &openapi3.RequestBodyRef{
		Value: documentAddRequest,
	}
	documentAdd.Responses = openapi3.NewResponses()
	documentAdd.Responses.Set("200", successResponse)
	documentAdd.Responses.Set("400", errorResponse)

	collectionUpdateWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionUpdateSchema))

	collectionUpdateWithResponse := openapi3.NewResponse().
		WithDescription("Update results").
		WithJSONSchemaRef(updateResultSchema)

	collectionUpdateWith := openapi3.NewOperation()
	collectionUpdateWith.OperationID = "collection_update_with_filter"
	collectionUpdateWith.Description = "Update document(s) in a collection"
	collectionUpdateWith.Tags = []string{"collection"}
	collectionUpdateWith.AddParameter(collectionNamePathParam)
	collectionUpdateWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionUpdateWithRequest,
	}
	collectionUpdateWith.AddResponse(200, collectionUpdateWithResponse)
	collectionUpdateWith.Responses.Set("400", errorResponse)

	collectionDeleteWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionDeleteSchema))

	collectionDeleteWithResponse := openapi3.NewResponse().
		WithDescription("Delete results").
		WithJSONSchemaRef(deleteResultSchema)

	collectionDeleteWith := openapi3.NewOperation()
	collectionDeleteWith.OperationID = "collection_delete_with_filter"
	collectionDeleteWith.Description = "Delete document(s) from a collection"
	collectionDeleteWith.Tags = []string{"collection"}
	collectionDeleteWith.AddParameter(collectionNamePathParam)
	collectionDeleteWith.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionDeleteWithRequest,
	}
	collectionDeleteWith.AddResponse(200, collectionDeleteWithResponse)
	collectionDeleteWith.Responses.Set("400", errorResponse)

	addIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(indexAddRequestSchema))
	addIndexResponse := openapi3.NewResponse().
		WithDescription("Index description").
		WithJSONSchemaRef(indexSchema)

	addIndex := openapi3.NewOperation()
	addIndex.OperationID = "index_add"
	addIndex.Description = "Add a secondary index"
	addIndex.Tags = []string{"index"}
	addIndex.AddParameter(collectionNamePathParam)
	addIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: addIndexRequest,
	}
	addIndex.AddResponse(200, addIndexResponse)
	addIndex.Responses.Set("400", errorResponse)

	indexArraySchema := openapi3.NewArraySchema()
	indexArraySchema.Items = indexSchema

	listIndexesResponse := openapi3.NewResponse().
		WithDescription("List of indexes").
		WithJSONSchema(indexArraySchema)

	listIndexes := openapi3.NewOperation()
	listIndexes.OperationID = "index_list"
	listIndexes.Description = "List secondary indexes"
	listIndexes.Tags = []string{"index"}
	listIndexes.AddParameter(collectionNamePathParam)
	listIndexes.AddResponse(200, listIndexesResponse)
	listIndexes.Responses.Set("400", errorResponse)

	indexPathParam := openapi3.NewPathParameter("index").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	deleteIndex := openapi3.NewOperation()
	deleteIndex.OperationID = "index_delete"
	deleteIndex.Description = "Delete a secondary index"
	deleteIndex.Tags = []string{"index"}
	deleteIndex.AddParameter(collectionNamePathParam)
	deleteIndex.AddParameter(indexPathParam)
	deleteIndex.Responses = openapi3.NewResponses()
	deleteIndex.Responses.Set("200", successResponse)
	deleteIndex.Responses.Set("400", errorResponse)

	documentIDPathParam := openapi3.NewPathParameter("docID").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentGetResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchemaRef(documentSchema)

	documentGet := openapi3.NewOperation()
	documentGet.Description = "Get a document by docID"
	documentGet.OperationID = "document_get"
	documentGet.Tags = []string{"document"}
	documentGet.AddParameter(collectionNamePathParam)
	documentGet.AddParameter(documentIDPathParam)
	documentGet.AddResponse(200, documentGetResponse)
	documentGet.Responses.Set("400", errorResponse)

	documentUpdate := openapi3.NewOperation()
	documentUpdate.Description = "Update a document by docID"
	documentUpdate.OperationID = "document_update"
	documentUpdate.Tags = []string{"document"}
	documentUpdate.AddParameter(collectionNamePathParam)
	documentUpdate.AddParameter(documentIDPathParam)
	documentUpdate.Responses = openapi3.NewResponses()
	documentUpdate.Responses.Set("200", successResponse)
	documentUpdate.Responses.Set("400", errorResponse)

	documentDelete := openapi3.NewOperation()
	documentDelete.Description = "Delete a document by docID"
	documentDelete.OperationID = "document_delete"
	documentDelete.Tags = []string{"document"}
	documentDelete.AddParameter(collectionNamePathParam)
	documentDelete.AddParameter(documentIDPathParam)
	documentDelete.Responses = openapi3.NewResponses()
	documentDelete.Responses.Set("200", successResponse)
	documentDelete.Responses.Set("400", errorResponse)

	addEncryptedIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(encryptedIndexAddRequestSchema))
	addEncryptedIndexResponse := openapi3.NewResponse().
		WithDescription("Encrypted index description").
		WithJSONSchemaRef(encryptedIndexSchema)

	addEncryptedIndex := openapi3.NewOperation()
	addEncryptedIndex.OperationID = "encrypted_index_add"
	addEncryptedIndex.Description = "Add an encrypted index"
	addEncryptedIndex.Tags = []string{"encrypted_index"}
	addEncryptedIndex.AddParameter(collectionNamePathParam)
	addEncryptedIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: addEncryptedIndexRequest,
	}
	addEncryptedIndex.AddResponse(200, addEncryptedIndexResponse)
	addEncryptedIndex.Responses.Set("400", errorResponse)

	encryptedIndexArraySchema := openapi3.NewArraySchema()
	encryptedIndexArraySchema.Items = encryptedIndexSchema

	getEncryptedIndexesResponse := openapi3.NewResponse().
		WithDescription("List of encrypted indexes").
		WithJSONSchema(encryptedIndexArraySchema)

	listEncryptedIndexes := openapi3.NewOperation()
	listEncryptedIndexes.OperationID = "encrypted_index_list"
	listEncryptedIndexes.Description = "List encrypted indexes"
	listEncryptedIndexes.Tags = []string{"encrypted_index"}
	listEncryptedIndexes.AddParameter(collectionNamePathParam)
	listEncryptedIndexes.AddResponse(200, getEncryptedIndexesResponse)
	listEncryptedIndexes.Responses.Set("400", errorResponse)

	fieldNamePathParam := openapi3.NewPathParameter("field").
		WithRequired(true).
		WithDescription("Field name of the encrypted index").
		WithSchema(openapi3.NewStringSchema())

	deleteEncryptedIndex := openapi3.NewOperation()
	deleteEncryptedIndex.OperationID = "encrypted_index_delete"
	deleteEncryptedIndex.Description = "Delete an encrypted index"
	deleteEncryptedIndex.Tags = []string{"encrypted_index"}
	deleteEncryptedIndex.AddParameter(collectionNamePathParam)
	deleteEncryptedIndex.AddParameter(fieldNamePathParam)
	deleteEncryptedIndex.Responses = openapi3.NewResponses()
	deleteEncryptedIndex.Responses.Set("200", successResponse)
	deleteEncryptedIndex.Responses.Set("400", errorResponse)

	truncate := openapi3.NewOperation()
	truncate.OperationID = "truncate"
	truncate.Description = "Truncate a collection, removing all document data within it from the server. " +
		"Does not propagate the deletion to other Defra nodes in the network."
	truncate.Tags = []string{"truncate"}
	truncate.AddParameter(collectionNamePathParam)
	truncate.Responses = openapi3.NewResponses()
	truncate.Responses.Set("200", successResponse)
	truncate.Responses.Set("400", errorResponse)

	router.AddRoute("/collections/{name}", http.MethodPost, documentAdd, h.AddDocument)
	router.AddRoute("/collections/{name}", http.MethodPatch, collectionUpdateWith, h.UpdateDocumentsWithFilter)
	router.AddRoute("/collections/{name}", http.MethodDelete, collectionDeleteWith, h.DeleteDocumentsWithFilter)
	router.AddRoute("/collections/{name}/indexes", http.MethodPost, addIndex, h.AddIndex)
	router.AddRoute("/collections/{name}/indexes", http.MethodGet, listIndexes, h.ListIndexes)
	router.AddRoute("/collections/{name}/indexes/{index}", http.MethodDelete, deleteIndex, h.DeleteIndex)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodPost, addEncryptedIndex,
		h.AddEncryptedIndex)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodGet, listEncryptedIndexes,
		h.ListEncryptedIndexes)
	router.AddRoute("/collections/{name}/encrypted-indexes/{field}", http.MethodDelete, deleteEncryptedIndex,
		h.DeleteEncryptedIndex)
	router.AddRoute("/collections/{name}/truncate", http.MethodDelete, truncate, h.Truncate)

	router.AddRoute("/collections/{name}/document/{docID}", http.MethodGet, documentGet, h.GetDocument)
	router.AddRoute("/collections/{name}/document/{docID}", http.MethodPatch, documentUpdate, h.UpdateDocument)
	router.AddRoute("/collections/{name}/document/{docID}", http.MethodDelete, documentDelete, h.DeleteDocument)
}
