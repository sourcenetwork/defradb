// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clitest

import (
	"encoding/json"
	"testing"

	"github.com/sourcenetwork/defradb/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexList_IfCollectionIsNotSpecified_ShouldReturnAllIndexes(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createCollection(t, conf, `type User { name: String }`)
	createCollection(t, conf, `type Product { name: String price: Int }`)
	createIndexOnField(t, conf, "User", "name", "")
	createIndexOnField(t, conf, "Product", "name", "")
	createIndexOnField(t, conf, "Product", "price", "")

	stdout, _ := runDefraCommand(t, conf, []string{"client", "index", "list"})
	nodeLog := stopDefra()

	var resp struct {
		Data struct {
			Collections map[string][]client.IndexDescription `json:"collections"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(stdout[0]), &resp)
	require.NoError(t, err)

	assert.Equal(t, len(resp.Data.Collections), 2)
	assert.Equal(t, len(resp.Data.Collections["User"]), 1)
	assert.Equal(t, len(resp.Data.Collections["Product"]), 2)

	assertNotContainsSubstring(t, stdout, "errors")
	assertNotContainsSubstring(t, nodeLog, "errors")
}

func TestIndexList_IfCollectionIsSpecified_ShouldReturnCollectionsIndexes(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)
	createIndexOnName(t, conf)

	createCollection(t, conf, `type Product { name: String price: Int }`)
	createIndexOnField(t, conf, "Product", "name", "")
	createIndexOnField(t, conf, "Product", "price", "")

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "list",
		"--collection", "User",
	})
	nodeLog := stopDefra()

	var resp struct {
		Data struct {
			Indexes []client.IndexDescription `json:"indexes"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(stdout[0]), &resp)
	require.NoError(t, err)

	expectedDesc := client.IndexDescription{Name: userColIndexOnNameFieldName, ID: 1, Fields: []client.IndexedFieldDescription{{Name: "name", Direction: client.Ascending}}}
	assert.Equal(t, 1, len(resp.Data.Indexes))
	assert.Equal(t, expectedDesc, resp.Data.Indexes[0])

	assertNotContainsSubstring(t, stdout, "errors")
	assertNotContainsSubstring(t, nodeLog, "errors")
}

func TestIndexList_IfInternalError_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "index", "list",
		"--collection", "User",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "errors")
}
