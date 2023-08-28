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

	"github.com/go-chi/chi/v5"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

type lensHandler struct{}

func (s *lensHandler) ReloadLenses(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	err := lens.ReloadLenses(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) SetMigration(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var cfg client.LensConfig
	if err := requestJSON(req, &cfg); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	err := lens.SetMigration(req.Context(), cfg)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *lensHandler) MigrateUp(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var src enumerable.Enumerable[map[string]any]
	if err := requestJSON(req, &src); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	result, err := lens.MigrateUp(req.Context(), src, chi.URLParam(req, "version"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (s *lensHandler) MigrateDown(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	var src enumerable.Enumerable[map[string]any]
	if err := requestJSON(req, &src); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	result, err := lens.MigrateDown(req.Context(), src, chi.URLParam(req, "version"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, result)
}

func (s *lensHandler) Config(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	cfgs, err := lens.Config(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, cfgs)
}

func (s *lensHandler) HasMigration(rw http.ResponseWriter, req *http.Request) {
	lens := req.Context().Value(lensContextKey).(client.LensRegistry)

	exists, err := lens.HasMigration(req.Context(), chi.URLParam(req, "version"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err.Error()})
		return
	}
	if !exists {
		responseJSON(rw, http.StatusBadRequest, errorResponse{"migration not found"})
		return
	}
	rw.WriteHeader(http.StatusOK)
}
