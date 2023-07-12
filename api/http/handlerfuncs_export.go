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
	"net/http"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/client"
)

func exportHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	showDeletedParam := req.URL.Query().Get("showDeleted")
	showDeleted := showDeletedParam == "true"
	colsList := req.URL.Query()["collections"]

	// map of collection name pointing to slices of documents.
	docs := make(map[string][]map[string]any)
	getDocs := func(rw http.ResponseWriter, req *http.Request, col client.Collection) (docs []map[string]any, ok bool) {
		keysCh, err := col.GetAllDocKeys(req.Context())
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return nil, false
		}

		for key := range keysCh {
			doc, err := col.Get(req.Context(), key.Key, showDeleted)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusInternalServerError)
				return nil, false
			}

			docM, err := doc.ToMap()
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusInternalServerError)
				return nil, false
			}
			delete(docM, "_key")
			newDoc, err := client.NewDocFromMap(docM)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusInternalServerError)
				return nil, false
			}
			// newKey is needed to let the user know what will be the key of the imported document.
			docM["_newKey"] = newDoc.Key().String()
			// NewDocFromMap removes the "_key" map item so we add it back.
			docM["_key"] = doc.Key().String()

			docs = append(docs, docM)
		}

		return docs, true
	}

	if len(colsList) == 0 {
		cols, err := db.GetAllCollections(req.Context())
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusInternalServerError)
			return
		}

		for _, col := range cols {
			var ok bool
			if docs[col.Schema().Name], ok = getDocs(rw, req, col); !ok {
				return
			}
		}
	} else {
		for _, colName := range colsList {
			col, err := db.GetCollectionByName(req.Context(), colName)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusBadRequest)
				return
			}

			var ok bool
			if docs[col.Schema().Name], ok = getDocs(rw, req, col); !ok {
				return
			}
		}
	}

	sendJSON(
		req.Context(),
		rw,
		DataResponse{docs},
		http.StatusOK,
	)
}

func importHandler(rw http.ResponseWriter, req *http.Request) {
	db, err := dbFromContext(req.Context())
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}

	cols := map[string][]map[string]any{}

	switch req.Header.Get("content-type") {
	case "application/octet-stream":
		err = cbor.NewDecoder(req.Body).Decode(&cols)
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusBadRequest)
			return
		}
	default:
		err = getJSON(req, &cols)
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusBadRequest)
			return
		}
	}

	txn, err := db.NewTxn(req.Context(), false)
	if err != nil {
		handleErr(req.Context(), rw, err, http.StatusInternalServerError)
		return
	}
	defer txn.Discard(req.Context())

	txnDB := db.WithTxn(txn)

	for colName, docs := range cols {
		col, err := txnDB.GetCollectionByName(req.Context(), colName)
		if err != nil {
			handleErr(req.Context(), rw, err, http.StatusBadRequest)
			return
		}
		for _, docMap := range docs {
			newkey, ok := docMap["_newKey"].(string)
			if !ok {
				handleErr(req.Context(), rw, ErrMissingNewKey, http.StatusBadRequest)
				return
			}
			key, err := client.NewDocKeyFromString(newkey)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusBadRequest)
				return
			}

			doc := client.NewDocWithKey(key)
			for k, v := range docMap {
				if k == "_key" || k == "_newKey" {
					continue
				}
				err := doc.Set(k, v)
				if err != nil {
					handleErr(req.Context(), rw, err, http.StatusBadRequest)
					return
				}
			}

			err = col.Create(req.Context(), doc)
			if err != nil {
				handleErr(req.Context(), rw, err, http.StatusBadRequest)
				return
			}
		}
	}
	err = txn.Commit(req.Context())
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
