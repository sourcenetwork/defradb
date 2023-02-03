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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinPathsWithBase(t *testing.T) {
	path, err := JoinPaths("http://localhost:9181", BlocksPath, "cid_of_some_sort")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http://localhost:9181"+BlocksPath+"/cid_of_some_sort", path.String())
}

func TestJoinPathsWithNoBase(t *testing.T) {
	_, err := JoinPaths("", BlocksPath, "cid_of_some_sort")
	assert.ErrorIs(t, ErrSchema, err)
}

func TestJoinPathsWithBaseWithoutHttpPrefix(t *testing.T) {
	_, err := JoinPaths("localhost:9181", BlocksPath, "cid_of_some_sort")
	assert.ErrorIs(t, ErrSchema, err)
}

func TestJoinPathsWithNoPaths(t *testing.T) {
	path, err := JoinPaths("http://localhost:9181")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http://localhost:9181", path.String())
}

func TestJoinPathsWithInvalidCharacter(t *testing.T) {
	_, err := JoinPaths("https://%gh&%ij")
	assert.Error(t, err)
}
