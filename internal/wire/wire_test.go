// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type registeredType struct{}

func init() {
	Register[registeredType]()
}

// TestRegistered checks that Registered returns a registered type and not an
// unregistered one.
func TestRegistered(t *testing.T) {
	var names []string
	for _, ty := range Registered() {
		names = append(names, ty.Name())
	}
	assert.Contains(t, names, "registeredType")
	assert.NotContains(t, names, "unregisteredType")
}
