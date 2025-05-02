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

import (
	"fmt"

	"github.com/stretchr/testify/require"
)

// TxnCommit executes the `client txn commit` command using the txn id at the given
// state-index.
type TxCommit struct {
	stateful

	TxnIndex int
}

var _ Action = (*TxCommit)(nil)

func (a *TxCommit) Execute() {
	args := []string{"client", "tx", "commit", fmt.Sprint(a.s.Txns[a.TxnIndex])}
	args = a.AppendDirections(args)

	err := execute(a.s.Ctx, args)
	require.NoError(a.s.T, err)
}
