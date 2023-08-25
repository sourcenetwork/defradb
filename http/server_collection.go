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
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type CollectionHandler struct{}

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

func (s *CollectionHandler) Create(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var body any
	if err := requestJSON(req, &body); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}

	switch t := body.(type) {
	case []map[string]any:
		var docList []*client.Document
		for _, docMap := range t {
			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
				return
			}
			docList = append(docList, doc)
		}
		if err := col.CreateMany(req.Context(), docList); err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		rw.WriteHeader(http.StatusOK)
	case map[string]any:
		doc, err := client.NewDocFromMap(t)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		if err := col.Create(req.Context(), doc); err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		rw.WriteHeader(http.StatusOK)
	default:
		responseJSON(rw, http.StatusBadRequest, H{"error": "invalid request body"})
	}
}

func (s *CollectionHandler) Save(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var docMap map[string]any
	if err := requestJSON(req, &docMap); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	doc, err := client.NewDocFromMap(docMap)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err = col.Save(req.Context(), doc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *CollectionHandler) DeleteWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionDeleteRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.DeleteWith(req.Context(), request.Filter)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		result, err := col.DeleteWith(req.Context(), docKey)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.DeleteWith(req.Context(), docKeys)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, H{"error": "invalid delete request"})
	}
}

func (s *CollectionHandler) UpdateWith(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var request CollectionUpdateRequest
	if err := requestJSON(req, &request); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}

	switch {
	case request.Filter != nil:
		result, err := col.UpdateWith(req.Context(), request.Filter, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Key != "":
		docKey, err := client.NewDocKeyFromString(request.Key)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		result, err := col.UpdateWith(req.Context(), docKey, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	case request.Keys != nil:
		var docKeys []client.DocKey
		for _, key := range request.Keys {
			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
				return
			}
			docKeys = append(docKeys, docKey)
		}
		result, err := col.UpdateWith(req.Context(), docKeys, request.Updater)
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, result)
	default:
		responseJSON(rw, http.StatusBadRequest, H{"error": "invalid update request"})
	}
}

func (s *CollectionHandler) Update(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var docMap map[string]any
	if err := requestJSON(req, &docMap); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	doc, err := client.NewDocFromMap(docMap)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	if doc.Key().String() != chi.URLParam(req, "key") {
		responseJSON(rw, http.StatusBadRequest, H{"error": "document key does not match"})
		return
	}
	err = col.Update(req.Context(), doc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *CollectionHandler) Delete(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	_, err = col.Delete(req.Context(), docKey)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *CollectionHandler) Get(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)
	showDeleted, _ := strconv.ParseBool(req.URL.Query().Get("deleted"))

	docKey, err := client.NewDocKeyFromString(chi.URLParam(req, "key"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	_, err = col.Get(req.Context(), docKey, showDeleted)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

type DocKeyResult struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

func (s *CollectionHandler) GetAllDocKeys(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, H{"error": "streaming not supported"})
		return
	}

	docKeyCh, err := col.GetAllDocKeys(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
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

func (s *CollectionHandler) CreateIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	var indexDesc client.IndexDescription
	if err := requestJSON(req, &indexDesc); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	index, err := col.CreateIndex(req.Context(), indexDesc)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, index)
}

func (s *CollectionHandler) GetIndexes(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	indexes, err := col.GetIndexes(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *CollectionHandler) DropIndex(rw http.ResponseWriter, req *http.Request) {
	col := req.Context().Value(colContextKey).(client.Collection)

	err := col.DropIndex(req.Context(), chi.URLParam(req, "index"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}
