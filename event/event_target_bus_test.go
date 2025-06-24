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

package event

import (
	"syscall/js"
	"testing"

	"github.com/sourcenetwork/goji"
	"github.com/stretchr/testify/suite"
)

func TestEventTargetBus(t *testing.T) {
	suite.Run(t, NewBusTestSuite(func(buffer int) Bus {
		value := goji.EventTarget.New()
		return NewEventTargetBus(js.Value(value), buffer)
	}))
}
