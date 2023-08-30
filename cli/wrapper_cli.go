// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/datastore"
)

type cliWrapper struct {
	address string
	txValue string
}

func newCliWrapper(address string) *cliWrapper {
	return &cliWrapper{
		address: address,
	}
}

func (w *cliWrapper) withTxn(tx datastore.Txn) *cliWrapper {
	return &cliWrapper{
		address: w.address,
		txValue: fmt.Sprintf("%d", tx.ID()),
	}
}

func (w *cliWrapper) execute(ctx context.Context, args []string) ([]byte, error) {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	if w.txValue != "" {
		args = append(args, "--tx", w.txValue)
	}

	cfg := config.DefaultConfig()
	cfg.API.Address = w.address

	cmd := NewDefraCommand(cfg)
	cmd.SetOut(&stdOut)
	cmd.SetErr(&stdErr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		return nil, err
	}
	if stdErr.Len() > 0 {
		return nil, fmt.Errorf("%s", stdErr.String())
	}
	return stdOut.Bytes(), nil
}
