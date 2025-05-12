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

	"github.com/stretchr/testify/assert"
)

func TestSetGlobal(t *testing.T) {
	SetGlobal()

	defradb := js.Global().Get("defradb")
	assert.Equal(t, defradb.Type(), js.TypeObject)
}
