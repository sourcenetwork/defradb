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
type unregisteredType struct{}
type localType struct{}

func init() {
	Register[registeredType]()
	MarkLocal[localType]()
}

// TestIsRegistered checks that Register records a type, MarkLocal does not add
// one to the wire set, and an unregistered type reads as absent.
func TestIsRegistered(t *testing.T) {
	assert.True(t, IsRegistered[registeredType]())
	assert.False(t, IsRegistered[unregisteredType]())
	assert.False(t, IsRegistered[localType](), "MarkLocal must not add to the wire set")
}

// TestRegisteredContainsRegistered checks the snapshot accessor returns the
// registered type and not a local or unregistered one.
func TestRegisteredContainsRegistered(t *testing.T) {
	var names []string
	for _, ty := range Registered() {
		names = append(names, ty.Name())
	}
	assert.Contains(t, names, "registeredType")
	assert.NotContains(t, names, "localType")
	assert.NotContains(t, names, "unregisteredType")
}
