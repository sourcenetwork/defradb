// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceManagerDefault(t *testing.T) {
	rcmr, err := buildResourceManager("")
	assert.Nil(t, rcmr)
	assert.Nil(t, err)
}

func TestResourceManagerLimited(t *testing.T) {
	rcmr, err := buildResourceManager(ResourceProfileLimited)
	assert.NoError(t, err)
	assert.NotNil(t, rcmr)

}

func TestResourceManagerServer(t *testing.T) {
	rcmr, err := buildResourceManager(ResourceProfileServer)
	assert.NoError(t, err)
	assert.NotNil(t, rcmr)
}

func TestResourceManagerUnknown(t *testing.T) {
	rcmr, err := buildResourceManager("unknown")
	assert.Error(t, err)
	assert.Nil(t, rcmr)
}
