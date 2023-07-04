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
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

func createIndexHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	var data map[string]string
	err = getJSON(req, &data)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	colNameArg := data["collection"]
	fieldsArg := data["fields"]
	indexNameArg := data["name"]

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
		simpleDataResponse("index", indexDesc),
		http.StatusOK,
	)
}

func dropIndexHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	var data map[string]string
	err = getJSON(req, &data)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	colNameArg := data["collection"]
	indexNameArg := data["name"]

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
		sendJSON(
			req.Context(),
			rw,
			simpleDataResponse("collections", indexesPerCol),
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
		sendJSON(
			req.Context(),
			rw,
			simpleDataResponse("indexes", indexes),
			http.StatusOK,
		)
	}
}
