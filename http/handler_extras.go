// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/event"
)

// extrasHandler contains additional http handlers not found in client interfaces.
type extrasHandler struct{}

func (s *extrasHandler) Purge(rw http.ResponseWriter, req *http.Request) {
	db := req.Context().Value(dbContextKey).(client.DB)
	rw.WriteHeader(http.StatusOK) // write the response before we restart to purge
	db.Events().Publish(event.NewMessage(event.PurgeName, nil))
}

func (h *extrasHandler) bindRoutes(router *Router) {
	errorResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/error",
	}
	successResponse := &openapi3.ResponseRef{
		Ref: "#/components/responses/success",
	}

	purge := openapi3.NewOperation()
	purge.Description = "Purge all persisted data and restart"
	purge.OperationID = "purge"
	purge.Responses = openapi3.NewResponses()
	purge.Responses.Set("200", successResponse)
	purge.Responses.Set("400", errorResponse)

	router.AddRoute("/purge", http.MethodPost, purge, h.Purge)
}
