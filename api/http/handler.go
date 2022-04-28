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
	"github.com/pkg/errors"
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

func getJSON(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return errors.Wrap(err, "unmarshall error")
	}
	return nil
}

func sendJSON(rw http.ResponseWriter, v interface{}, code int) {
	rw.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(v)
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("Error while encoding JSON: %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, `{"error": "Internal server error"}`)
		return
	}

	rw.WriteHeader(code)
	rw.Write(b)
}
