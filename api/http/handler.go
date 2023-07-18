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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/pkg/errors"

	"github.com/sourcenetwork/defradb/client"
)

type handler struct {
	db client.DB
	*chi.Mux

	// user configurable options
	options serverOptions
}

// context variables
type (
	ctxDB     struct{}
	ctxPeerID struct{}
)

// DataResponse is the GQL top level object holding data for the response payload.
type DataResponse struct {
	Data any `json:"data"`
}

// simpleDataResponse is a helper function that returns a DataResponse struct.
// Odd arguments are the keys and must be strings otherwise they are ignored.
// Even arguments are the values associated with the previous key.
// Odd arguments are also ignored if there are no following arguments.
func simpleDataResponse(args ...any) DataResponse {
	data := make(map[string]any)

	for i := 0; i < len(args); i += 2 {
		if len(args) >= i+2 {
			switch a := args[i].(type) {
			case string:
				data[a] = args[i+1]

			default:
				continue
			}
		}
	}

	return DataResponse{
		Data: data,
	}
}

// newHandler returns a handler with the router instantiated.
func newHandler(db client.DB, opts serverOptions) *handler {
	mux := chi.NewRouter()
	mux.Use(loggerMiddleware)

	if len(opts.allowedOrigins) != 0 {
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins: opts.allowedOrigins,
			AllowedMethods: []string{"GET", "POST", "PATCH", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
			MaxAge:         300,
		}))
	}

	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if opts.tls.HasValue() {
				rw.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			ctx := context.WithValue(req.Context(), ctxDB{}, db)
			if opts.peerID != "" {
				ctx = context.WithValue(ctx, ctxPeerID{}, opts.peerID)
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	})

	return setRoutes(&handler{
		Mux:     mux,
		db:      db,
		options: opts,
	})
}

func getJSON(req *http.Request, v any) error {
	err := json.NewDecoder(req.Body).Decode(v)
	if err != nil {
		return errors.Wrap(err, "unmarshal error")
	}
	return nil
}

func sendJSON(ctx context.Context, rw http.ResponseWriter, v any, code int) {
	rw.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(v)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("Error while encoding JSON: %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(rw, `{"error": "Internal server error"}`); err != nil {
			log.Error(ctx, err.Error())
		}
		return
	}

	rw.WriteHeader(code)
	if _, err = rw.Write(b); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Error(ctx, err.Error())
	}
}

func dbFromContext(ctx context.Context) (client.DB, error) {
	db, ok := ctx.Value(ctxDB{}).(client.DB)
	if !ok {
		return nil, ErrDatabaseNotAvailable
	}

	return db, nil
}
