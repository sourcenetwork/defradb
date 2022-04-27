// Copyright 2022 Democratized Data Foundation
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

func TestServer(t *testing.T) {
	// @TODO: maybe it would be worth doing something a bit more thorough

	// test with no config
	s := NewServer()
	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(":303000"))
	}
}
