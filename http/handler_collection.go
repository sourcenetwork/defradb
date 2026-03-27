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

type DeleteCollectionRequest struct {
	Filter any `json:"filter"`
}

type UpdateCollectionRequest struct {
	Filter  any    `json:"filter"`
	Updater string `json:"updater"`
}

func (h *collectionHandler) DeleteDocumentsWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request DeleteCollectionRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteOpt := options.WithIdentity(options.DeleteDocumentsWithFilter(), identity.FromContext(ctx))

	result, err := col.DeleteDocumentsWithFilter(ctx, request.Filter, deleteOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) UpdateDocumentsWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request UpdateCollectionRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	updateOpt := options.WithIdentity(options.UpdateDocumentsWithFilter(), identity.FromContext(ctx))

	result, err := col.UpdateDocumentsWithFilter(ctx, request.Filter, request.Updater, updateOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) NewIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	descWithoutID := client.NewIndexRequest{
		Name:   indexDesc.Name,
		Fields: indexDesc.Fields,
		Unique: indexDesc.Unique,
	}

	newIndexOpt := options.WithIdentity(options.NewCollectionIndex(), identity.FromContext(ctx))

	index, err := col.NewIndex(ctx, descWithoutID, newIndexOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
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
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}

	listIndexesOpt := options.WithIdentity(options.ListCollectionIndexes(), ident)

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

	deleteIndexOpt := options.WithIdentity(options.DeleteCollectionIndex(), identity.FromContext(ctx))

	err := col.DeleteIndex(ctx, chi.URLParam(req, "index"), deleteIndexOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) NewEncryptedIndex(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	var indexDesc client.EncryptedIndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	opts := options.WithIdentity(options.NewEncryptedIndex(), identity.FromContext(req.Context()))

	index, err := col.NewEncryptedIndex(req.Context(), indexDesc, opts)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (h *collectionHandler) ListEncryptedIndexes(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	opts := options.WithIdentity(options.ListCollectionEncryptedIndexes(), identity.FromContext(req.Context()))
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
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) Truncate(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	truncateOpt := options.WithIdentity(options.TruncateCollection(), identity.FromContext(ctx))

	err := col.Truncate(ctx, truncateOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
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
		Ref: "#/components/schemas/update_collection",
	}
	updateResultSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/update_result",
	}
	collectionDeleteSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/delete_collection",
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
	newIndexRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/new_index",
	}
	encryptedIndexSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/encrypted_index",
	}
	newEncryptedIndexRequestSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/new_encrypted_index",
	}

	collectionNamePathParam := openapi3.NewPathParameter("name").
		WithDescription("Collection name").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	documentArraySchema := openapi3.NewArraySchema()
	documentArraySchema.Items = documentSchema

	addDocumentBodySchema := openapi3.NewOneOfSchema()
	addDocumentBodySchema.OneOf = openapi3.SchemaRefs{
		documentSchema,
		openapi3.NewSchemaRef("", documentArraySchema),
	}

	addDocumentRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(addDocumentBodySchema))

	addDocument := openapi3.NewOperation()
	addDocument.OperationID = "add_document"
	addDocument.Description = "Add document(s) to a collection"
	addDocument.Tags = []string{"document"}
	addDocument.AddParameter(collectionNamePathParam)
	addDocument.RequestBody = &openapi3.RequestBodyRef{
		Value: addDocumentRequest,
	}
	addDocument.Responses = openapi3.NewResponses()
	addDocument.Responses.Set("200", successResponse)
	addDocument.Responses.Set("400", errorResponse)
	addDocument.Responses.Set("409", errorResponse)

	updateCollectionWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionUpdateSchema))

	updateCollectionWithResponse := openapi3.NewResponse().
		WithDescription("Update results").
		WithJSONSchemaRef(updateResultSchema)

	updateCollectionWith := openapi3.NewOperation()
	updateCollectionWith.OperationID = "update_collection_with_filter"
	updateCollectionWith.Description = "Update document(s) in a collection"
	updateCollectionWith.Tags = []string{"collection"}
	updateCollectionWith.AddParameter(collectionNamePathParam)
	updateCollectionWith.RequestBody = &openapi3.RequestBodyRef{
		Value: updateCollectionWithRequest,
	}
	updateCollectionWith.AddResponse(200, updateCollectionWithResponse)
	updateCollectionWith.Responses.Set("400", errorResponse)

	deleteCollectionWithRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(collectionDeleteSchema))

	deleteCollectionWithResponse := openapi3.NewResponse().
		WithDescription("Delete results").
		WithJSONSchemaRef(deleteResultSchema)

	deleteCollectionWith := openapi3.NewOperation()
	deleteCollectionWith.OperationID = "delete_collection_with_filter"
	deleteCollectionWith.Description = "Delete document(s) from a collection"
	deleteCollectionWith.Tags = []string{"collection"}
	deleteCollectionWith.AddParameter(collectionNamePathParam)
	deleteCollectionWith.RequestBody = &openapi3.RequestBodyRef{
		Value: deleteCollectionWithRequest,
	}
	deleteCollectionWith.AddResponse(200, deleteCollectionWithResponse)
	deleteCollectionWith.Responses.Set("400", errorResponse)

	newIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(newIndexRequestSchema))
	newIndexResponse := openapi3.NewResponse().
		WithDescription("Index description").
		WithJSONSchemaRef(indexSchema)

	newIndex := openapi3.NewOperation()
	newIndex.OperationID = "new_index"
	newIndex.Description = "Make a new secondary index"
	newIndex.Tags = []string{"index"}
	newIndex.AddParameter(collectionNamePathParam)
	newIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: newIndexRequest,
	}
	newIndex.AddResponse(200, newIndexResponse)
	newIndex.Responses.Set("400", errorResponse)
	newIndex.Responses.Set("409", errorResponse)

	indexArraySchema := openapi3.NewArraySchema()
	indexArraySchema.Items = indexSchema

	listIndexesResponse := openapi3.NewResponse().
		WithDescription("List of indexes").
		WithJSONSchema(indexArraySchema)

	listIndexes := openapi3.NewOperation()
	listIndexes.OperationID = "list_indexes"
	listIndexes.Description = "List secondary indexes"
	listIndexes.Tags = []string{"index"}
	listIndexes.AddParameter(collectionNamePathParam)
	listIndexes.AddResponse(200, listIndexesResponse)
	listIndexes.Responses.Set("400", errorResponse)

	indexPathParam := openapi3.NewPathParameter("index").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	deleteIndex := openapi3.NewOperation()
	deleteIndex.OperationID = "delete_index"
	deleteIndex.Description = "Delete a secondary index"
	deleteIndex.Tags = []string{"index"}
	deleteIndex.AddParameter(collectionNamePathParam)
	deleteIndex.AddParameter(indexPathParam)
	deleteIndex.Responses = openapi3.NewResponses()
	deleteIndex.Responses.Set("200", successResponse)
	deleteIndex.Responses.Set("400", errorResponse)
	deleteIndex.Responses.Set("404", errorResponse)

	documentIDPathParam := openapi3.NewPathParameter("docID").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	getDocumentResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchemaRef(documentSchema)

	getDocument := openapi3.NewOperation()
	getDocument.Description = "Get a document by docID"
	getDocument.OperationID = "get_document"
	getDocument.Tags = []string{"document"}
	getDocument.AddParameter(collectionNamePathParam)
	getDocument.AddParameter(documentIDPathParam)
	getDocument.AddResponse(200, getDocumentResponse)
	getDocument.Responses.Set("400", errorResponse)
	getDocument.Responses.Set("404", errorResponse)

	updateDocument := openapi3.NewOperation()
	updateDocument.Description = "Update a document by docID"
	updateDocument.OperationID = "update_document"
	updateDocument.Tags = []string{"document"}
	updateDocument.AddParameter(collectionNamePathParam)
	updateDocument.AddParameter(documentIDPathParam)
	updateDocument.Responses = openapi3.NewResponses()
	updateDocument.Responses.Set("200", successResponse)
	updateDocument.Responses.Set("400", errorResponse)
	updateDocument.Responses.Set("404", errorResponse)

	deleteDocument := openapi3.NewOperation()
	deleteDocument.Description = "Delete a document by docID"
	deleteDocument.OperationID = "delete_document"
	deleteDocument.Tags = []string{"document"}
	deleteDocument.AddParameter(collectionNamePathParam)
	deleteDocument.AddParameter(documentIDPathParam)
	deleteDocument.Responses = openapi3.NewResponses()
	deleteDocument.Responses.Set("200", successResponse)
	deleteDocument.Responses.Set("400", errorResponse)
	deleteDocument.Responses.Set("404", errorResponse)

	newEncryptedIndexRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchemaRef(newEncryptedIndexRequestSchema))
	newEncryptedIndexResponse := openapi3.NewResponse().
		WithDescription("Encrypted index description").
		WithJSONSchemaRef(encryptedIndexSchema)

	newEncryptedIndex := openapi3.NewOperation()
	newEncryptedIndex.OperationID = "new_encrypted_index"
	newEncryptedIndex.Description = "Make a new encrypted index"
	newEncryptedIndex.Tags = []string{"encrypted_index"}
	newEncryptedIndex.AddParameter(collectionNamePathParam)
	newEncryptedIndex.RequestBody = &openapi3.RequestBodyRef{
		Value: newEncryptedIndexRequest,
	}
	newEncryptedIndex.AddResponse(200, newEncryptedIndexResponse)
	newEncryptedIndex.Responses.Set("400", errorResponse)
	newEncryptedIndex.Responses.Set("409", errorResponse)

	encryptedIndexArraySchema := openapi3.NewArraySchema()
	encryptedIndexArraySchema.Items = encryptedIndexSchema

	listEncryptedIndexesResponse := openapi3.NewResponse().
		WithDescription("List of encrypted indexes").
		WithJSONSchema(encryptedIndexArraySchema)

	listEncryptedIndexes := openapi3.NewOperation()
	listEncryptedIndexes.OperationID = "list_encrypted_indexes"
	listEncryptedIndexes.Description = "List encrypted indexes"
	listEncryptedIndexes.Tags = []string{"encrypted_index"}
	listEncryptedIndexes.AddParameter(collectionNamePathParam)
	listEncryptedIndexes.AddResponse(200, listEncryptedIndexesResponse)
	listEncryptedIndexes.Responses.Set("400", errorResponse)

	fieldNamePathParam := openapi3.NewPathParameter("field").
		WithRequired(true).
		WithDescription("Field name of the encrypted index").
		WithSchema(openapi3.NewStringSchema())

	deleteEncryptedIndex := openapi3.NewOperation()
	deleteEncryptedIndex.OperationID = "delete_encrypted_index"
	deleteEncryptedIndex.Description = "Delete an encrypted index"
	deleteEncryptedIndex.Tags = []string{"encrypted_index"}
	deleteEncryptedIndex.AddParameter(collectionNamePathParam)
	deleteEncryptedIndex.AddParameter(fieldNamePathParam)
	deleteEncryptedIndex.Responses = openapi3.NewResponses()
	deleteEncryptedIndex.Responses.Set("200", successResponse)
	deleteEncryptedIndex.Responses.Set("400", errorResponse)
	deleteEncryptedIndex.Responses.Set("404", errorResponse)

	truncate := openapi3.NewOperation()
	truncate.OperationID = "truncate"
	truncate.Description = "Truncate a collection, removing all document data within it from the server. " +
		"Does not propagate the deletion to other Defra nodes in the network."
	truncate.Tags = []string{"truncate"}
	truncate.AddParameter(collectionNamePathParam)
	truncate.Responses = openapi3.NewResponses()
	truncate.Responses.Set("200", successResponse)
	truncate.Responses.Set("400", errorResponse)

	router.AddRoute("/collections/{name}", http.MethodPost, addDocument, h.AddDocument)
	router.AddRoute("/collections/{name}", http.MethodPatch, updateCollectionWith, h.UpdateDocumentsWithFilter)
	router.AddRoute("/collections/{name}", http.MethodDelete, deleteCollectionWith, h.DeleteDocumentsWithFilter)
	router.AddRoute("/collections/{name}/indexes", http.MethodPost, newIndex, h.NewIndex)
	router.AddRoute("/collections/{name}/indexes", http.MethodGet, listIndexes, h.ListIndexes)
	router.AddRoute("/collections/{name}/indexes/{index}", http.MethodDelete, deleteIndex, h.DeleteIndex)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodPost, newEncryptedIndex,
		h.NewEncryptedIndex)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodGet, listEncryptedIndexes,
		h.ListEncryptedIndexes)
	router.AddRoute("/collections/{name}/encrypted-indexes/{field}", http.MethodDelete, deleteEncryptedIndex,
		h.DeleteEncryptedIndex)
	router.AddRoute("/collections/{name}/truncate", http.MethodDelete, truncate, h.Truncate)

	router.AddRoute("/collections/{name}/document/{docID}", http.MethodGet, getDocument, h.GetDocument)
	router.AddRoute("/collections/{name}/document/{docID}", http.MethodPatch, updateDocument, h.UpdateDocument)
	router.AddRoute("/collections/{name}/document/{docID}", http.MethodDelete, deleteDocument, h.DeleteDocument)
}
