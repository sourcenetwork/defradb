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
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/internal/db"
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

	tx, ok := db.TryGetContextTxn(ctx)
	if ok {
		args = append(args, "--tx", fmt.Sprintf("%d", tx.ID()))
	}
	id := identity.FromContext(ctx)
	if id.HasValue() && id.Value().PrivateKey != nil {
		args = append(args, "--identity", hex.EncodeToString(id.Value().PrivateKey.Serialize()))
		args = append(args, "--source-hub-address", w.sourceHubAddress)
	}
	args = append(args, "--url", w.address)

	cmd := cli.NewDefraCommand()
	cmd.SetOut(stdOutWrite)
	cmd.SetErr(stdErrWrite)
	cmd.SetArgs(args)

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	go func() {
		err := cmd.Execute()
		stdOutWrite.CloseWithError(err)
		stdErrWrite.CloseWithError(err)
	}()

	return stdOutRead, stdErrRead, nil
}
