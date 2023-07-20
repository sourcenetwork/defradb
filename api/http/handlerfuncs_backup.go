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
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

func exportHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	cfg := &client.BackupConfig{}
	err = getJSON(req, cfg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	err = validateBackupConfig(req.Context(), cfg, db)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	err = db.BasicExport(req.Context(), cfg)
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

	cfg := &client.BackupConfig{}
	err = getJSON(req, cfg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	err = validateBackupConfig(req.Context(), cfg, db)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusBadRequest)
		return
	}

	err = db.BasicImport(req.Context(), cfg.Filepath)
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

func validateBackupConfig(ctx context.Context, cfg *client.BackupConfig, db client.DB) error {
	if !isValidPath(cfg.Filepath) {
		return errors.New("invalid file path")
	}

	if cfg.Format != "" && strings.ToLower(cfg.Format) != "json" {
		return errors.New("only JSON format is supported at the moment")
	}
	for _, colName := range cfg.Collections {
		_, err := db.GetCollectionByName(ctx, colName)
		if err != nil {
			return errors.Wrap("collection does not exist", err)
		}
	}
	return nil
}

func isValidPath(filepath string) bool {
	// if a file exists, return true
	if _, err := os.Stat(filepath); err == nil {
		return true
	}

	// if not, attempt to write to the path and if successful,
	// remove the file and return true
	var d []byte
	if err := os.WriteFile(filepath, d, 0o644); err == nil {
		os.Remove(filepath)
		return true
	}

	return false
}
