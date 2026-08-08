// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/identity"
)

const purgeDocIDsProperty = "docIDs"

type purgeDocIDsRequest struct {
	DocIDs       []string `json:"docIDs"`
	PruneHistory bool     `json:"pruneHistory"`
}

func (h *storeHandler) PurgeDocuments(rw http.ResponseWriter, req *http.Request) {
	db := mustGetContextClientDB(req)
	ctx := req.Context()
	txn, hasTxn := datastore.CtxTryGetClientTxn(ctx)

	var body purgeDocIDsRequest
	if err := requestJSON(req, &body); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if len(body.DocIDs) == 0 {
		responseJSON(rw, http.StatusBadRequest, errorResponse{
			NewErrMissingRequiredParameter(purgeDocIDsProperty),
		})
		return
	}

	docIDs := make([]client.DocID, 0, len(body.DocIDs))
	for _, raw := range body.DocIDs {
		docID, err := client.NewDocIDFromString(raw)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		docIDs = append(docIDs, docID)
	}

	purgeOpt := options.WithIdentity(options.PurgeDocuments(), identity.FromContext(ctx))
	collectionName := chi.URLParam(req, "name")
	var err error
	if hasTxn {
		err = txn.PurgeDocuments(ctx, collectionName, docIDs, body.PruneHistory, purgeOpt)
	} else {
		err = db.PurgeDocuments(ctx, collectionName, docIDs, body.PruneHistory, purgeOpt)
	}
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}

	rw.WriteHeader(http.StatusOK)
}
