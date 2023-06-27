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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type indexFieldResp struct {
	Name string `json:"name"`
}
type indexResp struct {
	Name   string           `json:"name"`
	Fields []indexFieldResp `json:"fields"`
}

func TestIndexList_IfCollectionIsNotSpecified_ShouldReturnAllIndexes(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createCollection(t, conf, `type User { Name: String }`)
	createCollection(t, conf, `type Product { Name: String Price: Int }`)
	createIndexOnField(t, conf, "User", "Name", "")
	createIndexOnField(t, conf, "Product", "Name", "")
	createIndexOnField(t, conf, "Product", "Price", "")

	stdout, _ := runDefraCommand(t, conf, []string{"client", "index", "list"})
	nodeLog := stopDefra()

	var resp struct {
		Data struct {
			Collections map[string][]indexResp `json:"collections"`
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

	createCollection(t, conf, `type Product { Name: String Price: Int }`)
	createIndexOnField(t, conf, "Product", "Name", "")
	createIndexOnField(t, conf, "Product", "Price", "")

	stdout, _ := runDefraCommand(t, conf, []string{"client", "index", "list",
		"--collection", "User"})
	nodeLog := stopDefra()

	var resp struct {
		Data struct {
			Indexes []indexResp `json:"indexes"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(stdout[0]), &resp)
	require.NoError(t, err)

	expectedResp := indexResp{Name: userColIndexOnNameFieldName, Fields: []indexFieldResp{{Name: "Name"}}}
	assert.Equal(t, 1, len(resp.Data.Indexes))
	assert.Equal(t, expectedResp, resp.Data.Indexes[0])

	assertNotContainsSubstring(t, stdout, "errors")
	assertNotContainsSubstring(t, nodeLog, "errors")
}

func TestIndexList_IfInternalError_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "index", "list",
		"--collection", "User"})
	stopDefra()

	assertContainsSubstring(t, stdout, "errors")
}
