// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallingFuncName(t *testing.T) {
	name := callingFuncName()
	assert.Equal(t, "TestCallingFuncName", name)
}

func TestCallingFuncPackage(t *testing.T) {
	name := callingPackageName()
	assert.Equal(t, "github.com/sourcenetwork/defradb/internal/metric", name)
}
