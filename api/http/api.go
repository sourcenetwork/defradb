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
	"io/ioutil"
	"log"
	"net/http"

	"github.com/multiformats/go-multihash"
	"github.com/sourcenetwork/defradb/client"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	dag "github.com/ipfs/go-merkledag"
)

type Server struct {
	db     client.DB
	router *chi.Mux
}

func NewServer(db client.DB) *Server {
	s := &Server{
		db: db,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database"))
		if err != nil {
			log.Printf("DefraDB HTTP API Welcome message writing failed: %v", err)
		}
	})

	r.Get("/ping", s.ping)
	r.Get("/dump", s.dump)
	r.Get("/blocks/get/{cid}", s.getBlock)
	r.Get("/graphql", s.execGQL)
	r.Post("/schema/load", s.loadSchema)
	s.router = r
	return s
}

func (s *Server) Listen(addr string) {
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Fatalln("Error: HTTP Listening and Serving Failed: ", err)
	}
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))
	if err != nil {
		log.Printf("Writing pong with HTTP failed: %v", err)
	}
}

func (s *Server) dump(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	s.db.PrintDump(ctx)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		log.Printf("Writing ok with HTTP failed: %v", err)
	}
}

func (s *Server) execGQL(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	query := r.URL.Query().Get("query")
	result := s.db.ExecQuery(ctx, query)

	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (s *Server) loadSchema(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var result client.QueryResult
	sdl, err := ioutil.ReadAll(r.Body)

	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Print(err) // Should this be `log.Fatal(err)` ??
		}
	}()

	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.db.AddSchema(ctx, string(sdl))
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]string{
		"result": "success",
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (s *Server) getBlock(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var result client.QueryResult
	cidStr := chi.URLParam(r, "cid")

	// try to parse CID
	c, err := cid.Decode(cidStr)
	if err != nil {
		// if we cant try to parse DSKeyToCID
		// return error if we still cant
		key := ds.NewKey(cidStr)
		var hash multihash.Multihash
		hash, err = dshelp.DsKeyToMultihash(key)
		if err != nil {
			result.Errors = []interface{}{err.Error()}
			result.Data = err.Error()

			err = json.NewEncoder(w).Encode(result)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c = cid.NewCidV1(cid.Raw, hash)
	}

	block, err := s.db.GetBlock(ctx, c)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = err.Error()

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buf, err := nd.MarshalJSON()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Println(string(buf))

	reg := corecrdt.LWWRegister{}
	delta, err := reg.DeltaDecode(nd)
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := delta.Marshal()
	if err != nil {
		result.Errors = []interface{}{err.Error()}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// var val interface{}
	// err = cbor.Unmarshal(delta.Value().([]byte), &val)
	// if err != nil {
	//   result.Errors = []interface{}{err.Error()}
	//   err = json.NewEncoder(w).Encode(result)
	//   if err != nil {
	//     http.Error(w, err.Error(), 500)
	//     return
	//   }
	//   w.WriteHeader(http.StatusBadRequest)
	//   return
	// }
	result.Data = map[string]interface{}{
		"block": string(buf),
		"delta": string(data),
		"val":   delta.Value(),
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	err = enc.Encode(result)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = nil

		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
