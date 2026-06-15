// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"syscall/js"
	"testing"

	"github.com/sourcenetwork/goji"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalAndOpen(t *testing.T) {
	SetGlobal()

	defradb := js.Global().Get("defradb")
	require.Equal(t, js.TypeObject, defradb.Type())

	opts := map[string]any{"Store": map[string]any{"Store": "memory"}}
	prom := defradb.Call("open", opts)
	res, err := goji.Await(goji.PromiseValue(prom))
	require.NoError(t, err)

	require.Len(t, res, 1)
	assert.Equal(t, js.TypeObject, res[0].Type())
}
