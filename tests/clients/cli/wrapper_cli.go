// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/internal/datastore"
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

	tx, ok := datastore.CtxTryGetClientTxn(ctx)
	if ok {
		args = append(args, "--tx", fmt.Sprintf("%d", tx.ID()))
	}
	if len(w.sourceHubAddress) > 0 {
		args = append(args, "--source-hub-address", w.sourceHubAddress)
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
