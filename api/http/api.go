package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sourcenetwork/defradb/client"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
		w.Write([]byte("Welcome to the DefraDB HTTP API. Use /graphql to send queries to the database"))
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
	http.ListenAndServe(addr, s.router)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) dump(w http.ResponseWriter, r *http.Request) {
	s.db.PrintDump()
	w.Write([]byte("ok"))
}

func (s *Server) execGQL(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	result := s.db.ExecQuery(query)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) loadSchema(w http.ResponseWriter, r *http.Request) {
	var result client.QueryResult
	sdl, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.db.LoadSchema(string(sdl))
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result.Data = map[string]string{
		"result": "success",
	}
	json.NewEncoder(w).Encode(result)
}

func (s *Server) getBlock(w http.ResponseWriter, r *http.Request) {
	var result client.QueryResult
	cidStr := chi.URLParam(r, "cid")
	fmt.Println(cidStr)

	key := ds.NewKey(cidStr)
	c, err := dshelp.DsKeyToCid(key)
	// c, err := cid.Decode(cidStr)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = err.Error()
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	block, err := s.db.GetBlock(c)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		result.Data = err.Error()
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buf, err := nd.MarshalJSON()
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Println(string(buf))

	reg := corecrdt.LWWRegister{}
	delta, err := reg.DeltaDecode(nd)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lwwdelta := delta.(*corecrdt.LWWRegDelta)
	var val interface{}
	err = cbor.Unmarshal(lwwdelta.Data, &val)
	if err != nil {
		result.Errors = []interface{}{err.Error()}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result.Data = map[string]interface{}{
		"block": string(buf),
		"delta": string(lwwdelta.Data),
		"val":   val,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	enc.Encode(result)
}
