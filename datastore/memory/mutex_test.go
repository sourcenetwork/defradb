// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package memory

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyMutex(t *testing.T) {
	km := newKeyMutex()

	km.lock("test")
	assert.Equal(t, int32(1), atomic.LoadInt32(km.keys["test"].queue))
	km.unlock("test")
	assert.Nil(t, km.keys["test"])

	km.rlock("test")
	km.rlock("test")
	assert.Equal(t, int32(2), atomic.LoadInt32(km.keys["test"].queue))
	km.runlock("test")
	km.runlock("test")
	assert.Nil(t, km.keys["test"])
}
