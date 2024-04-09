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
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

type lensHandler struct{}

func (s *lensHandler) ReloadLenses(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)

	err := db.LensRegistry().ReloadLenses(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) SetMigration(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)

	var request setMigrationRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	err := db.LensRegistry().SetMigration(req.Context(), request.CollectionID, request.Config)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) MigrateUp(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)

	var request migrateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	result, err := db.LensRegistry().MigrateUp(req.Context(), enumerable.New(request.Data), request.CollectionID)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	var value []map[string]any
	err = enumerable.ForEach(result, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, value)
}

func (s *lensHandler) MigrateDown(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)

	var request migrateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	result, err := db.LensRegistry().MigrateDown(req.Context(), enumerable.New(request.Data), request.CollectionID)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	var value []map[string]any
	err = enumerable.ForEach(result, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, value)
}

func (h *lensHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	migrateSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/migrate_request",
	}
	setMigrationSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/set_migration_request",
	}

	setMigrationRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(setMigrationSchema)

	setMigration := openapi3.NewOperation()
	setMigration.OperationID = "lens_registry_set_migration"
	setMigration.Description = "Add a new lens migration to registry"
	setMigration.Tags = []string{"lens"}
	setMigration.RequestBody = &openapi3.RequestBodyRef{
		Value: setMigrationRequest,
	}
	setMigration.Responses = openapi3.NewResponses()
	setMigration.Responses.Set("200", successResponse)
	setMigration.Responses.Set("400", errorResponse)

	reloadLenses := openapi3.NewOperation()
	reloadLenses.OperationID = "lens_registry_reload"
	reloadLenses.Description = "Reload lens migrations"
	reloadLenses.Tags = []string{"lens"}
	reloadLenses.Responses = openapi3.NewResponses()
	reloadLenses.Responses.Set("200", successResponse)
	reloadLenses.Responses.Set("400", errorResponse)

	versionPathParam := openapi3.NewPathParameter("version").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	migrateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(migrateSchema)

	migrateUp := openapi3.NewOperation()
	migrateUp.OperationID = "lens_registry_migrate_up"
	migrateUp.Description = "Migrate documents to a collection"
	migrateUp.Tags = []string{"lens"}
	migrateUp.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateUp.AddParameter(versionPathParam)
	migrateUp.Responses = openapi3.NewResponses()
	migrateUp.Responses.Set("200", successResponse)
	migrateUp.Responses.Set("400", errorResponse)

	migrateDown := openapi3.NewOperation()
	migrateDown.OperationID = "lens_registry_migrate_down"
	migrateDown.Description = "Migrate documents from a collection"
	migrateDown.Tags = []string{"lens"}
	migrateDown.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateDown.AddParameter(versionPathParam)
	migrateDown.Responses = openapi3.NewResponses()
	migrateDown.Responses.Set("200", successResponse)
	migrateDown.Responses.Set("400", errorResponse)

	router.AddRoute("/lens/registry", http.MethodPost, setMigration, h.SetMigration)
	router.AddRoute("/lens/registry/reload", http.MethodPost, reloadLenses, h.ReloadLenses)
	router.AddRoute("/lens/registry/{version}/up", http.MethodPost, migrateUp, h.MigrateUp)
	router.AddRoute("/lens/registry/{version}/down", http.MethodPost, migrateDown, h.MigrateDown)
}
