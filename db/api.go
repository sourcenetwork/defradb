// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
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
