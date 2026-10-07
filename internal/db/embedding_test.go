// Copyright 2026 Democratized Data Foundation
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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestSetEmbedding_PartialDocUpdate_FetchesMissingFields(t *testing.T) {
	ctx := context.Background()

	var receivedInputs []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req struct {
			Input any `json:"input"`
		}
		if err := json.Unmarshal(body, &req); err == nil {
			switch v := req.Input.(type) {
			case string:
				receivedInputs = append(receivedInputs, v)
			case []any:
				var strParts []string
				for _, part := range v {
					if s, ok := part.(string); ok {
						strParts = append(strParts, s)
					}
				}
				receivedInputs = append(receivedInputs, strings.Join(strParts, "\n"))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": [{"embedding": [0.1, 0.2, 0.3]}]}`))
	}))
	defer ts.Close()

	schema := fmt.Sprintf(`
		type User {
			name: String
			about: String
			v: [Float32!] @embedding(fields: ["name", "about"], provider: "openai", model: "text-embedding-3-small", url: "%s")
		}`, ts.URL)

	database, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer database.Close()

	_, err = database.AddCollection(ctx, schema)
	require.NoError(t, err)

	col, err := database.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// 1. Add initial doc with both fields
	initialDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Alice", "about": "Loves tea"}`), col.Version())
	require.NoError(t, err)
	err = col.AddDocument(ctx, initialDoc)
	require.NoError(t, err)

	require.Len(t, receivedInputs, 1)
	assert.Contains(t, receivedInputs[0], "Alice")
	assert.Contains(t, receivedInputs[0], "Loves tea")

	// 2. Save partial document with only "about" (name omitted from patch)
	patchDoc := client.NewDocWithoutDefaultsWithID(initialDoc.ID(), col.Version())
	require.NoError(t, patchDoc.Set(ctx, "about", "Loves coffee"))

	err = col.SaveDocument(ctx, patchDoc)
	require.NoError(t, err)

	// Verify that setEmbedding fetched "Alice" from the database and combined it with "Loves coffee"
	require.Len(t, receivedInputs, 2)
	assert.Contains(t, receivedInputs[1], "Alice")
	assert.Contains(t, receivedInputs[1], "Loves coffee")
}
