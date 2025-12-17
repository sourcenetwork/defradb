// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// primarily based on https://github.com/99designs/gqlgen/blob/master/graphql/handler/transport
package graphql

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
)

var log = corelog.NewLogger("gqltransport")

type Executor func(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult

type Transport interface {
	Supports(r *http.Request) bool
	Do(w http.ResponseWriter, r *http.Request, executer Executor)
	Methods() []string
}

type OpenAPITransport interface {
	Transport
	OpenAPI3Operation() *openapi3.Operation
}

const (
	acceptApplicationJson                = "application/json"
	acceptApplicationGraphqlResponseJson = "application/graphql-response+json"
)

func determineResponseContentType(
	r *http.Request,
) string {

	accept := r.Header.Get("Accept")
	if accept == "" {
		return acceptApplicationGraphqlResponseJson
	}

	for _, acceptPart := range strings.Split(accept, ",") {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(acceptPart))
		if err != nil {
			continue
		}
		switch mediaType {
		case "*/*", "application/*":
			return acceptApplicationGraphqlResponseJson
		case acceptApplicationJson:
			return acceptApplicationJson
		case "application/graphql-response+json":
			return acceptApplicationGraphqlResponseJson
		}
	}

	return acceptApplicationGraphqlResponseJson
}

func writeHeaders(w http.ResponseWriter, headers map[string][]string) {
	if len(headers) == 0 {
		headers = map[string][]string{
			// Stay with application/json (not application/graphql-response+json)
			// as it is not an actively supported protocol for now
			"Content-Type": {acceptApplicationJson},
		}
	}

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

// responseJSON writes a json response with the given status and data
// to the response writer. Any errors encountered will be logged.
func responseJSON(rw http.ResponseWriter, status int, data any) {
	rw.WriteHeader(status)
	err := json.NewEncoder(rw).Encode(data)
	if err != nil {
		log.ErrorE("failed to write response", err)
	}
}
