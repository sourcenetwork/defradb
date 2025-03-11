// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import "github.com/stretchr/testify/require"

// TxCreate executes the `client tx create` command and appends the returned transaction id
// to state.Txns.
type TxCreate struct {
	stateful
}

var _ Action = (*TxCreate)(nil)

func (a *TxCreate) Execute() {
	result, err := executeJson[map[string]any](a.s.Ctx, a.AppendDirections([]string{"client", "tx", "create"}))
	require.NoError(a.s.T, err)

	txId, ok := result["id"].(float64)
	require.True(a.s.T, ok)

	a.s.Txns = append(a.s.Txns, uint64(txId))
}
