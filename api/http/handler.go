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

	"github.com/go-chi/chi"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	dag "github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multihash"
	"github.com/sourcenetwork/defradb/client"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/logging"
)

type handler struct {
	db client.DB

	*chi.Mux
	*logger
}

// HandlerConfig holds the handler configurable parameters
type HandlerConfig struct {
	Logger logging.Logger
}

// NewHandler returns a handler with the router instantiated and configuration applied.
func NewHandler(db client.DB, c *HandlerConfig) *handler {
	h := &handler{
		db: db,
	}

	if c != nil {
		if c.Logger != nil {
			h.logger = withLogger(c.Logger)
		}

	} else {
		h.logger = defaultLogger()
	}

	h.setRoutes()

	return h
}

type context struct {
	res http.ResponseWriter
	req *http.Request
	db  client.DB
	log *logger
}

func (h *handler) handle(f func(*context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(&context{
			res: w,
			req: r,
			db:  h.db,
			log: h.logger,
		})
	}
}

func root(c *context) {
	_, err := c.res.Write(
		[]byte("Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database"),
	)
	if err != nil {
		c.log.ErrorE(c.req.Context(), "DefraDB HTTP API Welcome message writing failed", err)
	}
}

func ping(c *context) {
	_, err := c.res.Write([]byte("pong"))
	if err != nil {
		c.log.ErrorE(c.req.Context(), "Writing pong with HTTP failed", err)
	}
}

func dump(c *context) {
	c.db.PrintDump(c.req.Context())

	_, err := c.res.Write([]byte("ok"))
	if err != nil {
		c.log.ErrorE(c.req.Context(), "Writing ok with HTTP failed", err)
	}
}

func execGQL(c *context) {
	var query string
	if c.req.Method == "GET" {
		query = c.req.URL.Query().Get("query")
	} else {
		body, err := io.ReadAll(c.req.Body)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusBadRequest)
			return
		}
		query = string(body)
	}

	if query == "" {
		http.Error(c.res, "missing GraphQL query", http.StatusBadRequest)
		return
	}

	result := c.db.ExecQuery(c.req.Context(), query)

	err := json.NewEncoder(c.res).Encode(result)
	if err != nil {
		http.Error(c.res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadSchema(c *context) {
	var result client.QueryResult
	sdl, err := io.ReadAll(c.req.Body)

	defer func() {
		err = c.req.Body.Close()
		if err != nil {
			c.log.ErrorE(c.req.Context(), "Error on body close", err)
		}
	}()

	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.db.AddSchema(c.req.Context(), string(sdl))
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]string{
		"result": "success",
	}

	err = json.NewEncoder(c.res).Encode(result)
	if err != nil {
		http.Error(c.res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getBlock(c *context) {
	var result client.QueryResult
	cidStr := chi.URLParam(c.req, "cid")

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

			err = json.NewEncoder(c.res).Encode(result)
			if err != nil {
				http.Error(c.res, err.Error(), http.StatusInternalServerError)
				return
			}

			c.res.WriteHeader(http.StatusBadRequest)
			return
		}
		cID = cid.NewCidV1(cid.Raw, hash)
	}

	block, err := c.db.Blockstore().Get(c.req.Context(), cID)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = err.Error()

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}
	buf, err := nd.MarshalJSON()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	reg := corecrdt.LWWRegister{}
	delta, err := reg.DeltaDecode(nd)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := delta.Marshal()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]interface{}{
		"block": string(buf),
		"delta": string(data),
		"val":   delta.Value(),
	}

	enc := json.NewEncoder(c.res)
	enc.SetIndent("", "\t")
	err = enc.Encode(result)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = nil

		err := json.NewEncoder(c.res).Encode(result)
		if err != nil {
			http.Error(c.res, err.Error(), http.StatusInternalServerError)
			return
		}

		c.res.WriteHeader(http.StatusBadRequest)
		return
	}
}
