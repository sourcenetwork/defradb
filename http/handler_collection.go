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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/encryption"
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

func (h *collectionHandler) Add(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	ctx := req.Context()
	q := req.URL.Query()
	encConf := encryption.DocEncConfig{}
	if q.Get(docEncryptParam) == "true" {
		encConf.IsDocEncrypted = true
	}
	if q.Get(docEncryptFieldsParam) != "" {
		encConf.EncryptedFields = strings.Split(q.Get(docEncryptFieldsParam), ",")
	}

	addOpt := options.WithIdentity(
		options.CollectionAdd().
			SetEncryptDoc(encConf.IsDocEncrypted).
			SetEncryptedFields(encConf.EncryptedFields),
		identity.FromContext(ctx),
	)

	switch {
	case client.IsJSONArray(data):
		docList, err := client.NewDocsFromJSON(ctx, data, col.Version())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}

		if err := col.AddMany(ctx, docList, addOpt); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		doc, err := client.NewDocFromJSON(ctx, data, col.Version())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		if err := col.Add(ctx, doc, addOpt); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *collectionHandler) DeleteWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteOpt := options.WithIdentity(options.CollectionDeleteWithFilter(), identity.FromContext(ctx))

	result, err := col.DeleteWithFilter(ctx, request.Filter, deleteOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) UpdateWithFilter(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	updateOpt := options.WithIdentity(options.CollectionUpdateWithFilter(), identity.FromContext(ctx))

	result, err := col.UpdateWithFilter(ctx, request.Filter, request.Updater, updateOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (h *collectionHandler) Update(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	col := mustGetContextClientCollection(req)

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	getOpt := options.WithIdentity(
		options.CollectionGet().SetShowDeleted(true),
		identity.FromContext(ctx),
	)

	doc, err := col.Get(ctx, docID, getOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
		return
	}

	patch, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if err := doc.SetWithJSON(ctx, patch); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	updateOpt := options.WithIdentity(options.CollectionUpdate(), identity.FromContext(ctx))

	err = col.Update(ctx, doc, updateOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) Delete(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteOpt := options.WithIdentity(options.CollectionDelete(), identity.FromContext(ctx))

	_, err = col.Delete(ctx, docID, deleteOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) Get(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("show_deleted"))

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	getOpt := options.WithIdentity(
		options.CollectionGet().SetShowDeleted(showDeleted),
		identity.FromContext(ctx),
	)

	doc, err := col.Get(ctx, docID, getOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
		return
	}

	docMap, err := doc.ToMap()
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docMap)
}

type DocIDResult struct {
	DocID string `json:"docID"`
	Error string `json:"error"`
}

func (h *collectionHandler) GetAllDocIDs(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	getAllOpt := options.WithIdentity(options.CollectionGetAllDocIDs(), identity.FromContext(ctx))

	docIDsResult, err := col.GetAllDocIDs(ctx, getAllOpt)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for docID := range docIDsResult {
		results := &DocIDResult{
			DocID: docID.ID.String(),
		}
		if docID.Err != nil {
			results.Error = docID.Err.Error()
		}
		data, err := json.Marshal(results)
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(rw, "data: %s\n\n", data)
		if err != nil {
			return
		}
		flusher.Flush()
	}
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

	collectionAddSchema := openapi3.NewOneOfSchema()
	collectionAddSchema.OneOf = openapi3.SchemaRefs{
		documentSchema,
		openapi3.NewSchemaRef("", documentArraySchema),
	}

	collectionAddRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(collectionAddSchema))

	collectionAdd := openapi3.NewOperation()
	collectionAdd.OperationID = "collection_add"
	collectionAdd.Description = "Add document(s) to a collection"
	collectionAdd.Tags = []string{"collection"}
	collectionAdd.AddParameter(collectionNamePathParam)
	collectionAdd.RequestBody = &openapi3.RequestBodyRef{
		Value: collectionAddRequest,
	}
	collectionAdd.Responses = openapi3.NewResponses()
	collectionAdd.Responses.Set("200", successResponse)
	collectionAdd.Responses.Set("400", errorResponse)

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

	collectionGetResponse := openapi3.NewResponse().
		WithDescription("Document value").
		WithJSONSchemaRef(documentSchema)

	collectionGet := openapi3.NewOperation()
	collectionGet.Description = "Get a document by docID"
	collectionGet.OperationID = "collection_get"
	collectionGet.Tags = []string{"collection"}
	collectionGet.AddParameter(collectionNamePathParam)
	collectionGet.AddParameter(documentIDPathParam)
	collectionGet.AddResponse(200, collectionGetResponse)
	collectionGet.Responses.Set("400", errorResponse)

	collectionUpdate := openapi3.NewOperation()
	collectionUpdate.Description = "Update a document by docID"
	collectionUpdate.OperationID = "collection_update"
	collectionUpdate.Tags = []string{"collection"}
	collectionUpdate.AddParameter(collectionNamePathParam)
	collectionUpdate.AddParameter(documentIDPathParam)
	collectionUpdate.Responses = openapi3.NewResponses()
	collectionUpdate.Responses.Set("200", successResponse)
	collectionUpdate.Responses.Set("400", errorResponse)

	collectionDelete := openapi3.NewOperation()
	collectionDelete.Description = "Delete a document by docID"
	collectionDelete.OperationID = "collection_delete"
	collectionDelete.Tags = []string{"collection"}
	collectionDelete.AddParameter(collectionNamePathParam)
	collectionDelete.AddParameter(documentIDPathParam)
	collectionDelete.Responses = openapi3.NewResponses()
	collectionDelete.Responses.Set("200", successResponse)
	collectionDelete.Responses.Set("400", errorResponse)

	collectionKeys := openapi3.NewOperation()
	collectionKeys.AddParameter(collectionNamePathParam)
	collectionKeys.Description = "Get all document IDs"
	collectionKeys.OperationID = "collection_keys"
	collectionKeys.Tags = []string{"collection"}
	collectionKeys.Responses = openapi3.NewResponses()
	collectionKeys.Responses.Set("200", successResponse)
	collectionKeys.Responses.Set("400", errorResponse)

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

	router.AddRoute("/collections/{name}", http.MethodGet, collectionKeys, h.GetAllDocIDs)
	router.AddRoute("/collections/{name}", http.MethodPost, collectionAdd, h.Add)
	router.AddRoute("/collections/{name}", http.MethodPatch, collectionUpdateWith, h.UpdateWithFilter)
	router.AddRoute("/collections/{name}", http.MethodDelete, collectionDeleteWith, h.DeleteWithFilter)
	router.AddRoute("/collections/{name}/indexes", http.MethodPost, addIndex, h.AddIndex)
	router.AddRoute("/collections/{name}/indexes", http.MethodGet, listIndexes, h.ListIndexes)
	router.AddRoute("/collections/{name}/indexes/{index}", http.MethodDelete, deleteIndex, h.DeleteIndex)
	router.AddRoute("/collections/{name}/{docID}", http.MethodGet, collectionGet, h.Get)
	router.AddRoute("/collections/{name}/{docID}", http.MethodPatch, collectionUpdate, h.Update)
	router.AddRoute("/collections/{name}/{docID}", http.MethodDelete, collectionDelete, h.Delete)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodPost, addEncryptedIndex,
		h.AddEncryptedIndex)
	router.AddRoute("/collections/{name}/encrypted-indexes", http.MethodGet, listEncryptedIndexes,
		h.ListEncryptedIndexes)
	router.AddRoute("/collections/{name}/encrypted-indexes/{field}", http.MethodDelete, deleteEncryptedIndex,
		h.DeleteEncryptedIndex)
	router.AddRoute("/collections/{name}/truncate", http.MethodDelete, truncate, h.Truncate)
}
