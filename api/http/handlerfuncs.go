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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
	dshelp "github.com/ipfs/boxo/datastore/dshelp"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"

	"github.com/sourcenetwork/defradb/client"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
)

const (
	contentTypeJSON           = "application/json"
	contentTypeGraphQL        = "application/graphql"
	contentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

func rootHandler(rw http.ResponseWriter, req *http.Request) {
	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse(
			"response", "Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database."+
				" Read the documentation at https://docs.source.network/.",
		),
		http.StatusOK,
	)
}

func pingHandler(rw http.ResponseWriter, req *http.Request) {
	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("response", "pong"),
		http.StatusOK,
	)
}

func dumpHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = db.PrintDump(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("response", "ok"),
		http.StatusOK,
	)
}

type gqlRequest struct {
	Request string `json:"query"`
}

func execGQLHandler(rw http.ResponseWriter, req *http.Request) {
	request := req.URL.Query().Get("query")
	if request == "" {
		// extract the media type from the content-type header
		contentType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		// mime.ParseMediaType will return an error (mime: no media type)
		// if there is no media type set (i.e. application/json).
		// This however is not a failing condition as not setting the content-type header
		// should still make for a valid request and hit our default switch case.
		if err != nil && err.Error() != "mime: no media type" {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}

		switch contentType {
		case contentTypeJSON:
			gqlReq := gqlRequest{}

			err := getJSON(req, &gqlReq)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusBadRequest)
				return
			}

			request = gqlReq.Request

		case contentTypeFormURLEncoded:
			handleErr(
				req.Context(),
				rw,
				ErrFormNotSupported,
				http.StatusBadRequest,
			)
			return

		case contentTypeGraphQL:
			fallthrough

		default:
			if req.Body == nil {
				handleErr(req.Context(), rw, ErrBodyEmpty, http.StatusBadRequest)
				return
			}
			body, err := readWithLimit(req.Body, rw)
			if err != nil {
				handleErr(req.Context(), rw, errors.WithStack(err), http.StatusInternalServerError)
				return
			}
			request = string(body)
		}
	}

	// if at this point request is still empty, return an error
	if request == "" {
		handleErr(req.Context(), rw, ErrMissingGQLRequest, http.StatusBadRequest)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	result := db.ExecRequest(req.Context(), request)

	if result.Pub != nil {
		subscriptionHandler(result.Pub, rw, req)
		return
	}

	sendJSON(req.Context(), rw, newGQLResult(result.GQL), http.StatusOK)
}

type fieldResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Internal bool   `json:"internal"`
}

type collectionResponse struct {
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	VersionID string          `json:"version_id"`
	Fields    []fieldResponse `json:"fields,omitempty"`
}

func listSchemaHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	cols, err := db.GetAllCollections(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	colResp := make([]collectionResponse, len(cols))
	for i, col := range cols {
		var fields []fieldResponse
		for _, field := range col.Schema().Fields {
			fieldRes := fieldResponse{
				ID:       field.ID.String(),
				Name:     field.Name,
				Internal: field.IsInternal(),
			}
			if field.IsObjectArray() {
				fieldRes.Kind = fmt.Sprintf("[%s]", field.Schema)
			} else if field.IsObject() {
				fieldRes.Kind = field.Schema
			} else {
				fieldRes.Kind = field.Kind.String()
			}
			fields = append(fields, fieldRes)
		}
		colResp[i] = collectionResponse{
			Name:      col.Name(),
			ID:        col.SchemaID(),
			VersionID: col.Schema().VersionID,
			Fields:    fields,
		}
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("collections", colResp),
		http.StatusOK,
	)
}

func loadSchemaHandler(rw http.ResponseWriter, req *http.Request) {
	sdl, err := readWithLimit(req.Body, rw)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	colDescs, err := db.AddSchema(req.Context(), string(sdl))
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	colResp := make([]collectionResponse, len(colDescs))
	for i, desc := range colDescs {
		col, err := db.GetCollectionByName(req.Context(), desc.Name)
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}
		colResp[i] = collectionResponse{
			Name:      col.Name(),
			ID:        col.SchemaID(),
			VersionID: col.Schema().VersionID,
		}
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("result", "success", "collections", colResp),
		http.StatusOK,
	)
}

func patchSchemaHandler(rw http.ResponseWriter, req *http.Request) {
	patch, err := readWithLimit(req.Body, rw)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = db.PatchSchema(req.Context(), string(patch))
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

func setMigrationHandler(rw http.ResponseWriter, req *http.Request) {
	cfgStr, err := readWithLimit(req.Body, rw)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	txn, err := db.NewTxn(req.Context(), false)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	var cfg client.LensConfig
	err = json.Unmarshal(cfgStr, &cfg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = db.LensRegistry().SetMigration(req.Context(), cfg)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	err = txn.Commit(req.Context())
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

func getMigrationHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	cfgs, err := db.LensRegistry().Config(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse("configuration", cfgs),
		http.StatusOK,
	)
}

func getBlockHandler(rw http.ResponseWriter, req *http.Request) {
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
			handleErr(req.Context(), rw, err, http.StatusBadRequest)
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
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	buf, err := nd.MarshalJSON()
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	reg := corecrdt.LWWRegister{}
	delta, err := reg.DeltaDecode(nd)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	data, err := delta.Marshal()
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse(
			"block", string(buf),
			"delta", string(data),
			"val", delta.Value(),
		),
		http.StatusOK,
	)
}

func peerIDHandler(rw http.ResponseWriter, req *http.Request) {
	peerID, ok := req.Context().Value(ctxPeerID{}).(string)
	if !ok || peerID == "" {
		handleErr(req.Context(), rw, ErrPeerIdUnavailable, http.StatusNotFound)
		return
	}

	sendJSON(
		req.Context(),
		rw,
		simpleDataResponse(
			"peerID", peerID,
		),
		http.StatusOK,
	)
}

func subscriptionHandler(pub *events.Publisher[events.Update], rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		handleErr(req.Context(), rw, ErrStreamingUnsupported, http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-req.Context().Done():
			pub.Unsubscribe()
			return
		case s, open := <-pub.Stream():
			if !open {
				return
			}
			b, err := json.Marshal(s)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(rw, "data: %s\n\n", b)
			flusher.Flush()
		}
	}
}

// maxBytes is an arbitrary limit to prevent unbounded message bodies being sent and read.
const maxBytes int64 = 100 * (1 << (10 * 2)) // 100MB

// readWithLimit reads from the reader until either EoF or the maximum number of bytes have been read.
func readWithLimit(reader io.ReadCloser, rw http.ResponseWriter) ([]byte, error) {
	reader = http.MaxBytesReader(rw, reader, maxBytes)

	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
