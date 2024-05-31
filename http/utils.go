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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/badger/v4"
)

func requestJSON(req *http.Request, out any) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// responseJSON writes a json response with the given status and data
// to the response writer. Any errors encountered will be logged.
func responseJSON(rw http.ResponseWriter, status int, data any) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)

	err := json.NewEncoder(rw).Encode(data)
	if err != nil {
		log.ErrorE("failed to write response", err)
	}
}

func parseError(msg any) error {
	switch msg {
	case client.ErrDocumentNotFoundOrNotAuthorized.Error():
		return client.ErrDocumentNotFoundOrNotAuthorized
	case badger.ErrTxnConflict.Error():
		return badger.ErrTxnConflict
	default:
		return fmt.Errorf("%s", msg)
	}
}
