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
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type ccipHandler struct{}

type CCIPRequest struct {
	Sender string `json:"sender"`
	Data   string `json:"data"`
}

type CCIPResponse struct {
	Data string `json:"data"`
}

// ExecCCIP handles GraphQL over Cross Chain Interoperability Protocol requests.
func (c *ccipHandler) ExecCCIP(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var ccipReq CCIPRequest
	switch req.Method {
	case http.MethodGet:
		ccipReq.Sender = chi.URLParam(req, "sender")
		ccipReq.Data = chi.URLParam(req, "data")
	case http.MethodPost:
		if err := requestJSON(req, &ccipReq); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
	}

	data, err := hex.DecodeString(strings.TrimPrefix(ccipReq.Data, "0x"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	var request GraphQLRequest
	if err := json.Unmarshal(data, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	result := store.ExecRequest(req.Context(), request.Query)
	if result.Pub != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}
	resultJSON, err := json.Marshal(result.GQL)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	resultHex := "0x" + hex.EncodeToString(resultJSON)
	responseJSON(rw, http.StatusOK, CCIPResponse{Data: resultHex})
}
