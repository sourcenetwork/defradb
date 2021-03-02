package db

import (
	api "github.com/sourcenetwork/defradb/api/http"
)

func (db *DB) Listen() {
	db.log.Infof("Running HTTP API at http://%s. Try it out at > curl http://%s/graphql", db.options.Address, db.options.Address)

	s := api.NewServer(db)
	s.Listen(db.options.Address)
}

// func (db *DB) handlePing(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("pong"))
// }

// func (db *DB) handleGraphqlReq(w http.ResponseWriter, r *http.Request) {
// 	query := r.URL.Query().Get("query")
// 	result := db.ExecQuery(query)
// 	json.NewEncoder(w).Encode(result)
// }
