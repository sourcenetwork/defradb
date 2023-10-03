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

	"github.com/sourcenetwork/defradb/client"
)

type p2pHandler struct{}

func (s *p2pHandler) PeerInfo(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	responseJSON(rw, http.StatusOK, p2p.PeerInfo())
}

func (s *p2pHandler) SetReplicator(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.SetReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) DeleteReplicator(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	var rep client.Replicator
	if err := requestJSON(req, &rep); err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	err := p2p.DeleteReplicator(req.Context(), rep)
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) GetAllReplicators(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	reps, err := p2p.GetAllReplicators(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, reps)
}

func (s *p2pHandler) AddP2PCollection(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	err := p2p.AddP2PCollection(req.Context(), chi.URLParam(req, "id"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) RemoveP2PCollection(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	err := p2p.RemoveP2PCollection(req.Context(), chi.URLParam(req, "id"))
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (s *p2pHandler) GetAllP2PCollections(rw http.ResponseWriter, req *http.Request) {
	p2p, ok := req.Context().Value(dbContextKey).(client.P2P)
	if !ok {
		responseJSON(rw, http.StatusBadRequest, errorResponse{ErrP2PDisabled})
		return
	}
	cols, err := p2p.GetAllP2PCollections(req.Context())
	if err != nil {
		responseJSON(rw, http.StatusBadRequest, errorResponse{err})
		return
	}
	responseJSON(rw, http.StatusOK, cols)
}
