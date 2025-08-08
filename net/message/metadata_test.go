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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetaData_AllMessageMethods_ShouldSucceed(t *testing.T) {
	metaData := MetaData{}
	// set all fields
	metaData.SetErrMessage("some error")
	metaData.SetMessageID("some message ID")
	metaData.SetSenderID("some peer ID")
	metaData.SetPubkey([]byte("pubkey"))
	metaData.SetSignature([]byte("signature"))
	metaData.SetVersion()

	// get all fields and assert values
	require.Equal(t, "some error", metaData.GetErrMessage())
	require.Equal(t, "some message ID", metaData.GetMessageID())
	require.Equal(t, "some peer ID", metaData.GetSenderID())
	require.Equal(t, []byte("pubkey"), metaData.GetPubkey())
	require.Equal(t, []byte("signature"), metaData.GetSignature())
	require.Equal(t, messageVersion, metaData.GetVersion())
}
