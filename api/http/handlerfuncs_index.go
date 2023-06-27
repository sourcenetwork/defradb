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
	"net/http"
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

type indexFieldResponse struct {
	Name      string                `json:"name"`
	Direction client.IndexDirection `json:"direction"`
}

type indexResponse struct {
	Name   string               `json:"name"`
	ID     uint32               `json:"id"`
	Fields []indexFieldResponse `json:"fields"`
}

func indexDescToResponse(desc client.IndexDescription) indexResponse {
	indexResp := indexResponse{
		Name: desc.Name,
		ID:   desc.ID,
	}

	for _, field := range desc.Fields {
		indexResp.Fields = append(indexResp.Fields, indexFieldResponse{
			Name:      field.Name,
			Direction: field.Direction,
		})
	}

	return indexResp
}

func createIndexHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = req.ParseForm()
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	colNameArg := req.Form.Get("collection")
	fieldsArg := req.Form.Get("fields")
	indexNameArg := req.Form.Get("name")

	col, err := db.GetCollectionByName(req.Context(), colNameArg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	fields := strings.Split(fieldsArg, ",")
	fieldDescriptions := make([]client.IndexedFieldDescription, 0, len(fields))
	for _, field := range fields {
		fieldDescriptions = append(fieldDescriptions, client.IndexedFieldDescription{Name: field})
	}
	indexDesc := client.IndexDescription{
		Name:   indexNameArg,
		Fields: fieldDescriptions,
	}
	indexDesc, err = col.CreateIndex(req.Context(), indexDesc)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("index", indexDescToResponse(indexDesc)),
		http.StatusOK,
	)
}

func dropIndexHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = req.ParseForm()
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	colNameArg := req.Form.Get("collection")
	indexNameArg := req.Form.Get("name")

	col, err := db.GetCollectionByName(req.Context(), colNameArg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = col.DropIndex(req.Context(), indexNameArg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("result", "success"),
		http.StatusOK,
	)
}

func listIndexHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	queryParams := req.URL.Query()
	collectionParam := queryParams.Get("collection")

	if collectionParam == "" {
		indexesPerCol, err := db.GetAllIndexes(req.Context())
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}
		type collectionIndexes struct {
			Collections map[client.CollectionName][]indexResponse `json:"collections"`
		}
		var resp collectionIndexes
		resp.Collections = make(map[client.CollectionName][]indexResponse)
		for colName, indexes := range indexesPerCol {
			for _, index := range indexes {
				resp.Collections[colName] = append(
					resp.Collections[colName],
					indexDescToResponse(index),
				)
			}
		}
		sendJSON(
			req.Context(),
			rw,
			struct {
				Data collectionIndexes `json:"data"`
			}{Data: resp},
			http.StatusOK,
		)
	} else {
		col, err := db.GetCollectionByName(req.Context(), collectionParam)
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}
		indexes, err := col.GetIndexes(req.Context())
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}
		var resp []indexResponse
		for _, index := range indexes {
			resp = append(resp, indexDescToResponse(index))
		}
		sendJSON(
			req.Context(),
			rw,
			simpleDataResponse("indexes", resp),
			http.StatusOK,
		)
	}
}
