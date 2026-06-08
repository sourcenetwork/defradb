// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReceive_StreamLargerThanMax_ReturnsErrMessageTooLarge(t *testing.T) {
	stream := bytes.NewReader(make([]byte, maxMessageSize+1))
	err := Receive(stream, "some peer ID", nil, &MetaData{})
	require.ErrorIs(t, err, ErrMessageTooLarge)
}
