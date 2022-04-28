// Copyright 2022 Democratized Data Foundation
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sourcenetwork/defradb/client"
)

type Handler struct {
	db client.DB
	*chi.Mux
}

type ctxKey string

// newHandler returns a handler with the router instantiated.
func newHandler(db client.DB) *Handler {
	return setRoutes(&Handler{db: db})
}

func (h *Handler) handle(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), ctxKey("DB"), h.db)
		f(rw, req.WithContext(ctx))
	}
}

func sendJSON(rw http.ResponseWriter, v interface{}, code int) {
	rw.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(v)
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("Error while encoding JSON: %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(rw, `{"error": "Internal server error"}`); err != nil {
			log.Error(context.Background(), err.Error())
		}
		return
	}

	rw.WriteHeader(code)
	if _, err = rw.Write(b); err != nil {
		log.Error(context.Background(), err.Error())
	}
}
