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
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	dag "github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/sourcenetwork/defradb/client"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
)

const (
	contentTypeJSON           = "application/json"
	contentTypeGraphQL        = "application/graphql"
	contentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

func rootHandler(rw http.ResponseWriter, req *http.Request) {
	_, err := rw.Write(
		[]byte("Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database"),
	)
	if err != nil {
		handleErr(
			req.Context(),
			rw,
			errors.WithMessage(
				err,
				"DefraDB HTTP API Welcome message writing failed",
			),
			http.StatusInternalServerError,
		)
	}
}

func pingHandler(rw http.ResponseWriter, req *http.Request) {
	_, err := rw.Write([]byte("pong"))
	if err != nil {
		handleErr(
			req.Context(),
			rw,
			errors.WithMessage(
				err,
				"Writing pong with HTTP failed",
			),
			http.StatusInternalServerError,
		)
	}
}

func dumpHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	db.PrintDump(req.Context())

	_, err = rw.Write([]byte("ok"))
	if err != nil {
		handleErr(
			req.Context(),
			rw,
			errors.WithMessage(
				err,
				"Writing ok with HTTP failed",
			),
			http.StatusInternalServerError,
		)
	}
}

func execGQLHandler(rw http.ResponseWriter, req *http.Request) {
	query := req.URL.Query().Get("query")

	if query == "" {
		switch req.Header.Get("Content-Type") {
		case contentTypeJSON:
			handleErr(
				req.Context(),
				rw,
				errors.New("content type application/json not yet supported"),
				http.StatusBadRequest,
			)
			return

		case contentTypeFormURLEncoded:
			handleErr(
				req.Context(),
				rw,
				errors.New("content type application/x-www-form-urlencoded not yet supported"),
				http.StatusBadRequest,
			)
			return

		case contentTypeGraphQL:
			fallthrough

		default:
			body, err := io.ReadAll(req.Body)
			if err != nil {
				handleErr(req.Context(), rw, errors.WithStack(err), http.StatusBadRequest)
				return
			}
			query = string(body)
		}
	}

	// if at this point query is still empty, return an error
	if query == "" {
		handleErr(req.Context(), rw, errors.New("missing GraphQL query"), http.StatusBadRequest)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	result := db.ExecQuery(req.Context(), query)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		handleErr(req.Context(), rw, errors.WithStack(err), http.StatusBadRequest)
		return
	}
}

func loadSchemaHandler(rw http.ResponseWriter, req *http.Request) {
	var result client.QueryResult
	sdl, err := io.ReadAll(req.Body)

	defer func() {
		err = req.Body.Close()
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
		}
	}()

	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	err = db.AddSchema(req.Context(), string(sdl))
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]string{
		"result": "success",
	}

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
		return
	}
}

func getBlockHandler(rw http.ResponseWriter, req *http.Request) {
	var result client.QueryResult
	cidStr := chi.URLParam(req, "cid")

	// try to parse CID
	cID, err := cid.Decode(cidStr)
	if err != nil {
		// If we can't try to parse DSKeyToCID
		// return error if we still can't
		key := ds.NewKey(cidStr)
		var hash multihash.Multihash
		hash, err = dshelp.DsKeyToMultihash(key)
		if err != nil {
			result.Errors = []interface{}{err.Error()}
			result.Data = err.Error()

			err = json.NewEncoder(rw).Encode(result)
			if err != nil {
				handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
				return
			}

			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		cID = cid.NewCidV1(cid.Raw, hash)
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	block, err := db.Blockstore().Get(req.Context(), cID)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = err.Error()

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	buf, err := nd.MarshalJSON()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	reg := corecrdt.LWWRegister{}
	delta, err := reg.DeltaDecode(nd)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := delta.Marshal()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]interface{}{
		"block": string(buf),
		"delta": string(data),
		"val":   delta.Value(),
	}

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "\t")
	err = enc.Encode(result)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = nil

		err := json.NewEncoder(rw).Encode(result)
		if err != nil {
			handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}
