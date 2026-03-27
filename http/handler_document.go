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
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func (h *collectionHandler) AddDocument(rw http.ResponseWriter, req *http.Request) {
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
		options.AddDocument().
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

		if err := col.AddManyDocuments(ctx, docList, addOpt); err != nil {
			responseJSON(rw, httpStatusFromError(err), errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		doc, err := client.NewDocFromJSON(ctx, data, col.Version())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		if err := col.AddDocument(ctx, doc, addOpt); err != nil {
			responseJSON(rw, httpStatusFromError(err), errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func (h *collectionHandler) UpdateDocument(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	col := mustGetContextClientCollection(req)

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	getOpt := options.WithIdentity(
		options.GetDocument().SetShowDeleted(true),
		identity.FromContext(ctx),
	)

	doc, err := col.GetDocument(ctx, docID, getOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusNotFound, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
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

	updateOpt := options.WithIdentity(options.UpdateDocument(), identity.FromContext(ctx))

	err = col.UpdateDocument(ctx, doc, updateOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) DeleteDocument(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	deleteOpt := options.WithIdentity(options.DeleteDocument(), identity.FromContext(ctx))

	_, err = col.DeleteDocument(ctx, docID, deleteOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h *collectionHandler) GetDocument(rw http.ResponseWriter, req *http.Request) {
	col := mustGetContextClientCollection(req)
	ctx := req.Context()
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("show_deleted"))

	docID, err := client.NewDocIDFromString(chi.URLParam(req, "docID"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	getOpt := options.WithIdentity(
		options.GetDocument().SetShowDeleted(showDeleted),
		identity.FromContext(ctx),
	)

	doc, err := col.GetDocument(ctx, docID, getOpt)
	if err != nil {
		responseJSON(rw, httpStatusFromError(err), errorResponse{err})
		return
	}

	if doc == nil {
		responseJSON(rw, http.StatusNotFound, errorResponse{client.ErrDocumentNotFoundOrNotAuthorized})
		return
	}

	docMap, err := doc.ToMap()
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docMap)
}
