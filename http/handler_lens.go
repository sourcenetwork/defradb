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
	"github.com/go-chi/chi/v5"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

type lensHandler struct{}

func (s *lensHandler) ReloadLenses(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	err := lens.ReloadLenses(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) SetMigration(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var cfg client.LensConfig
	if err := requestJSON(req, &cfg); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := lens.SetMigration(req.Context(), cfg)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) MigrateUp(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var src []map[string]any
	if err := requestJSON(req, &src); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	result, err := lens.MigrateUp(req.Context(), enumerable.New(src), chi.URLParam(req, "version"))
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
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var src []map[string]any
	if err := requestJSON(req, &src); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	result, err := lens.MigrateDown(req.Context(), enumerable.New(src), chi.URLParam(req, "version"))
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

func (s *lensHandler) Config(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	cfgs, err := lens.Config(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cfgs)
}

func (s *lensHandler) HasMigration(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	exists, err := lens.HasMigration(req.Context(), chi.URLParam(req, "version"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if !exists {
		responseJSON(rw, http.StatusNotFound, errorResponse{ErrMigrationNotFound})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *lensHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}
	documentSchema := &openapi3.SchemaRef{
		Ref: "#/components/schemas/document",
	}

	lensConfigSchema := openapi3.NewSchemaRef("#/components/schemas/lens_config", nil)
	lensConfigArraySchema := openapi3.NewArraySchema()
	lensConfigArraySchema.Items = lensConfigSchema

	lensConfigResponse := openapi3.NewResponse().
		WithDescription("Lens configurations").
		WithJSONSchema(lensConfigArraySchema)

	lensConfig := openapi3.NewOperation()
	lensConfig.OperationID = "lens_config"
	lensConfig.Description = "List lens migrations"
	lensConfig.Tags = []string{"lens"}
	lensConfig.AddResponse(200, lensConfigResponse)
	lensConfig.Responses.Set("400", errorResponse)

	setMigrationRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(lensConfigSchema)

	setMigration := openapi3.NewOperation()
	setMigration.OperationID = "lens_set_migration"
	setMigration.Description = "Add a new lens migration"
	setMigration.Tags = []string{"lens"}
	setMigration.RequestBody = &openapi3.RequestBodyRef{
		Value: setMigrationRequest,
	}
	setMigration.Responses = openapi3.NewResponses()
	setMigration.Responses.Set("200", successResponse)
	setMigration.Responses.Set("400", errorResponse)

	reloadLenses := openapi3.NewOperation()
	reloadLenses.OperationID = "lens_reload"
	reloadLenses.Description = "Reload lens migrations"
	reloadLenses.Tags = []string{"lens"}
	reloadLenses.Responses = openapi3.NewResponses()
	reloadLenses.Responses.Set("200", successResponse)
	reloadLenses.Responses.Set("400", errorResponse)

	versionPathParam := openapi3.NewPathParameter("version").
		WithRequired(true).
		WithSchema(openapi3.NewStringSchema())

	hasMigration := openapi3.NewOperation()
	hasMigration.OperationID = "lens_has_migration"
	hasMigration.Description = "Check if a migration exists"
	hasMigration.Tags = []string{"lens"}
	hasMigration.AddParameter(versionPathParam)
	hasMigration.Responses = openapi3.NewResponses()
	hasMigration.Responses.Set("200", successResponse)
	hasMigration.Responses.Set("400", errorResponse)

	migrateSchema := openapi3.NewArraySchema()
	migrateSchema.Items = documentSchema
	migrateRequest := openapi3.NewRequestBody().
		WithRequired(true).
		WithContent(openapi3.NewContentWithJSONSchema(migrateSchema))

	migrateUp := openapi3.NewOperation()
	migrateUp.OperationID = "lens_migrate_up"
	migrateUp.Description = "Migrate documents to a schema version"
	migrateUp.Tags = []string{"lens"}
	migrateUp.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateUp.AddParameter(versionPathParam)
	migrateUp.Responses = openapi3.NewResponses()
	migrateUp.Responses.Set("200", successResponse)
	migrateUp.Responses.Set("400", errorResponse)

	migrateDown := openapi3.NewOperation()
	migrateDown.OperationID = "lens_migrate_down"
	migrateDown.Description = "Migrate documents from a schema version"
	migrateDown.Tags = []string{"lens"}
	migrateDown.RequestBody = &openapi3.RequestBodyRef{
		Value: migrateRequest,
	}
	migrateDown.AddParameter(versionPathParam)
	migrateDown.Responses = openapi3.NewResponses()
	migrateDown.Responses.Set("200", successResponse)
	migrateDown.Responses.Set("400", errorResponse)

	router.AddRoute("/lens", http.MethodGet, lensConfig, h.Config)
	router.AddRoute("/lens", http.MethodPost, setMigration, h.SetMigration)
	router.AddRoute("/lens/reload", http.MethodPost, reloadLenses, h.ReloadLenses)
	router.AddRoute("/lens/{version}", http.MethodGet, hasMigration, h.HasMigration)
	router.AddRoute("/lens/{version}/up", http.MethodPost, migrateUp, h.MigrateUp)
	router.AddRoute("/lens/{version}/down", http.MethodPost, migrateDown, h.MigrateDown)
}
