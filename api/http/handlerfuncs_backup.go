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

	"github.com/sourcenetwork/defradb/client"
)

func exportHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	showDeleted := req.URL.Query().Get("showDeleted") == "true"
	colsList := req.URL.Query()["collections"]
	filepath := req.URL.Query().Get("filepath")
	pretty := req.URL.Query().Get("pretty") == "true"

	err = db.BasicExport(req.Context(), &client.BackupConfig{
		Filepath:    filepath,
		ShowDeleted: showDeleted,
		Collections: colsList,
		Pretty:      pretty,
	})
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

func importHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	data := map[string]string{}

	err = getJSON(req, &data)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	err = db.BasicImport(req.Context(), data["filepath"])
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("result", "success"),
		http.StatusOK,
	)
}
