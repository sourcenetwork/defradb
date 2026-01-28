// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package event

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestChannelBus(t *testing.T) {
	suite.Run(t, NewBusTestSuite(func(buffer int) Bus {
		return NewChannelBus(0, buffer)
	}))
}
