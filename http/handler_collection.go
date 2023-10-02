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

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type collectionHandler struct{}

type CollectionDeleteRequest struct {
	Key    string   `json:"key"`
	Keys   []string `json:"keys"`
	Filter any      `json:"filter"`
}

type CollectionUpdateRequest struct {
	Key     string   `json:"key"`
	Keys    []string `json:"keys"`
	Filter  any      `json:"filter"`
	Updater string   `json:"updater"`
}

func (s *collectionHandler) Create(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var body any
	if err := requestJSON(req, &body); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch t := body.(type) {
	case []any:
		var docList []*client.Document
		for _, v := range t {
			docMap, ok := v.(map[string]any)
			if !ok {
				responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
				return
			}
			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docList = append(docList, doc)
		}
		if err := col.CreateMany(req.Context(), docList); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	case map[string]any:
		doc, err := client.NewDocFromMap(t)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		if err := col.Create(req.Context(), doc); err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) DeleteWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.DeleteWith(req.Context(), request.Filter)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		result, err := col.DeleteWith(req.Context(), docKey)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.DeleteWith(req.Context(), docKeys)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) UpdateWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.UpdateWith(req.Context(), request.Filter, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		result, err := col.UpdateWith(req.Context(), docKey, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, errorResponse{err})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.UpdateWith(req.Context(), docKeys, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, errorResponse{err})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrInvalidRequestBody})
	}
}

func (s *collectionHandler) Update(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	doc, err := col.Get(req.Context(), docKey, true)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	patch, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	if err := doc.SetWithJSON(patch); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err = col.Update(req.Context(), doc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Delete(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	_, err = col.Delete(req.Context(), docKey)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *collectionHandler) Get(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("show_deleted"))

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	doc, err := col.Get(req.Context(), docKey, showDeleted)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	docMap, err := doc.ToMap()
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, docMap)
}

type DocKeyResult struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

func (s *collectionHandler) GetAllDocKeys(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrStreamingNotSupported})
		return
	}

	docKeyCh, err := col.GetAllDocKeys(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for docKey := range docKeyCh {
		results := &DocKeyResult{
			Key: docKey.Key.String(),
		}
		if docKey.Err != nil {
			results.Error = docKey.Err.Error()
		}
		data, err := json.Marshal(results)
		if err != nil {
			return
		}
		fmt.Fprintf(rw, "data: %s\n\n", data)
		flusher.Flush()
	}
}

func (s *collectionHandler) CreateIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	index, err := col.CreateIndex(req.Context(), indexDesc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (s *collectionHandler) GetIndexes(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	indexes, err := col.GetIndexes(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *collectionHandler) DropIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	err := col.DropIndex(req.Context(), chi.URLParam(req, "index"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}
