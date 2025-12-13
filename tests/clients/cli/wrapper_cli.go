// Copyright 2025 Democratized Data Foundation
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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli"
)

type cliWrapper struct {
	address          string
	sourceHubAddress string
}

func newCliWrapper(address string, sourceHubAddress string) *cliWrapper {
	return &cliWrapper{
		address:          strings.TrimPrefix(address, "http://"),
		sourceHubAddress: sourceHubAddress,
	}
}

func (w *cliWrapper) execute(ctx context.Context, args []string) ([]byte, error) {
	stdOut, stdErr, err := w.executeStream(ctx, args)
	if err != nil {
		return nil, err
	}
	stdOutData, err := io.ReadAll(stdOut)
	if err != nil {
		return nil, err
	}
	stdErrData, err := io.ReadAll(stdErr)
	if err != nil {
		return nil, err
	}
	if len(stdErrData) != 0 {
		return nil, fmt.Errorf("%s", stdErrData)
	}
	return stdOutData, nil
}

func (w *cliWrapper) executeStream(ctx context.Context, args []string) (io.ReadCloser, io.ReadCloser, error) {
	stdOutRead, stdOutWrite := io.Pipe()
	stdErrRead, stdErrWrite := io.Pipe()

	txId, ok := ctxTryGetClientTxnId(ctx)
	if ok {
		args = append(args, "--tx", fmt.Sprintf("%d", txId))
	}
	id := identity.FromContext(ctx)
	if id.HasValue() {
		if fullIdent, ok := id.Value().(identity.FullIdentity); ok && fullIdent.PrivateKey() != nil {
			args = append(args, "--identity", fullIdent.PrivateKey().String())
			args = append(args, "--source-hub-address", w.sourceHubAddress)
		}
	}
	args = append(args, "--url", w.address)

	cmd := cli.NewDefraCommand(ctx)
	cmd.SetOut(stdOutWrite)
	cmd.SetErr(stdErrWrite)
	cmd.SetArgs(args)

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	go func() {
		err := cmd.Execute()
		_ = stdOutWrite.CloseWithError(err)
		_ = stdErrWrite.CloseWithError(err)
	}()

	return stdOutRead, stdErrRead, nil
}

type txnKey struct{}

// CtxTryGetClientTxn returns a client transaction and a bool indicating if the
// txn was retrieved from the given context.
func ctxTryGetClientTxnId(ctx context.Context) (uint64, bool) {
	txn, ok := ctx.Value(txnKey{}).(uint64)
	return txn, ok
}

// CtxSetFromClientTxn returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func ctxSetFromClientTxnId(ctx context.Context, txnId uint64) context.Context {
	return context.WithValue(ctx, txnKey{}, txnId)
}
