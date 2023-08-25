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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sourcenetwork/defradb/client"
)

type GraphQLRequest struct {
	Query string `json:"query"`
}

type GraphQLResponse struct {
	Errors []string `json:"errors,omitempty"`
	Data   any      `json:"data"`
}

func (res *GraphQLResponse) UnmarshalJSON(data []byte) error {
	// decode numbers to json.Number
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return err
	}

	// fix errors type to match tests
	switch t := out["errors"].(type) {
	case []any:
		var errors []string
		for _, v := range t {
			errors = append(errors, v.(string))
		}
		res.Errors = errors
	default:
		res.Errors = nil
	}

	// fix data type to match tests
	switch t := out["data"].(type) {
	case []any:
		var fixed []map[string]any
		for _, v := range t {
			fixed = append(fixed, v.(map[string]any))
		}
		res.Data = fixed
	case map[string]any:
		res.Data = t
	default:
		res.Data = []map[string]any{}
	}

	return nil
}

type StoreHandler struct{}

func (s *StoreHandler) SetReplicator(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err := store.SetReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) DeleteReplicator(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err := store.DeleteReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) GetAllReplicators(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	reps, err := store.GetAllReplicators(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, reps)
}

func (s *StoreHandler) AddP2PCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	err := store.AddP2PCollection(req.Context(), chi.URLParam(req, "id"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) RemoveP2PCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	err := store.RemoveP2PCollection(req.Context(), chi.URLParam(req, "id"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) GetAllP2PCollections(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	cols, err := store.GetAllP2PCollections(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (s *StoreHandler) BasicImport(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err := store.BasicImport(req.Context(), config.Filepath)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) BasicExport(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var config client.BackupConfig
	if err := requestJSON(req, &config); err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err := store.BasicExport(req.Context(), &config)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) AddSchema(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	schema, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	cols, err := store.AddSchema(req.Context(), string(schema))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}

func (s *StoreHandler) PatchSchema(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	patch, err := io.ReadAll(req.Body)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	err = store.PatchSchema(req.Context(), string(patch))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *StoreHandler) GetCollection(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	switch {
	case req.URL.Query().Has("name"):
		col, err := store.GetCollectionByName(req.Context(), req.URL.Query().Get("name"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, col.Description())
	case req.URL.Query().Has("schema_id"):
		col, err := store.GetCollectionBySchemaID(req.Context(), req.URL.Query().Get("schema_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, col.Description())
	case req.URL.Query().Has("version_id"):
		col, err := store.GetCollectionByVersionID(req.Context(), req.URL.Query().Get("version_id"))
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		responseJSON(rw, http.StatusOK, col.Description())
	default:
		cols, err := store.GetAllCollections(req.Context())
		if err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
		colDesc := make([]client.CollectionDescription, len(cols))
		for i, col := range cols {
			colDesc[i] = col.Description()
		}
		responseJSON(rw, http.StatusOK, colDesc)
	}
}

func (s *StoreHandler) GetAllIndexes(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	indexes, err := store.GetAllIndexes(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
		return
	}
	responseJSON(rw, http.StatusOK, indexes)
}

func (s *StoreHandler) ExecRequest(rw http.ResponseWriter, req *http.Request) {
	store := req.Context().Value(storeContextKey).(client.Store)

	var request GraphQLRequest
	switch {
	case req.URL.Query().Get("query") != "":
		request.Query = req.URL.Query().Get("query")
	case req.Body != nil:
		if err := requestJSON(req, &request); err != nil {
			responseJSON(rw, http.StatusBadRequest, H{"error": err.Error()})
			return
		}
	default:
		responseJSON(rw, http.StatusBadRequest, H{"error": "missing request"})
		return
	}
	result := store.ExecRequest(req.Context(), request.Query)

	var errors []string
	for _, err := range result.GQL.Errors {
		errors = append(errors, err.Error())
	}
	if result.Pub == nil {
		responseJSON(rw, http.StatusOK, H{"data": result.GQL.Data, "errors": errors})
		return
	}
	flusher, ok := rw.(http.Flusher)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, H{"error": "streaming not supported"})
		return
	}

	rw.Header().Add("Content-Type", "text/event-stream")
	rw.Header().Add("Cache-Control", "no-cache")
	rw.Header().Add("Connection", "keep-alive")

	rw.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-req.Context().Done():
			return
		case item, open := <-result.Pub.Stream():
			if !open {
				return
			}
			data, err := json.Marshal(item)
			if err != nil {
				return
			}
			fmt.Fprintf(rw, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
