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
)

type errorResponse struct {
	Error string `json:"error"`
}

func parseErrorResponse(data []byte) error {
	var res errorResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return fmt.Errorf("%s", data)
	}
	return fmt.Errorf(res.Error)
}

func parseResponse(res *http.Response) error {
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return parseErrorResponse(data)
	}
	return nil
}

func parseJsonResponse(res *http.Response, out any) error {
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return parseErrorResponse(data)
	}
	return json.Unmarshal(data, &out)
}
