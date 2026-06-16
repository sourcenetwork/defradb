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
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/node"
)

type StartCli struct {
	stateful
	inlineArgs    []string
	expectedError error
	useDisk       bool
}

var _ Action = (*StartCli)(nil)

// Start with additional CLI arguments, and the expected an error (for example, if you
// are testing with flags that should cause the command to fail)
func StartWithArgsE(args []string, expectedErr error) *StartCli {
	cli := &StartCli{
		inlineArgs:    args,
		expectedError: expectedErr,
	}
	return cli
}

// Start with additional CLI arguments, without an expected error (the command should
// cause the service to start successfully)
func StartWithArgs(args []string) *StartCli {
	cli := &StartCli{
		inlineArgs:    args,
		expectedError: nil,
	}
	return cli
}

// The minimal Start command will call StartWithArgs using no additional arguments,
// and will not expect an error. This will cause Execute() to continue until the service
// has been successfully started.
func Start() *StartCli {
	return StartWithArgs([]string{})
}

// StartDisk starts the node using disk-based (Badger) storage. This allows testing
// options that behave differently in disk mode, such as ChunkSize (which is always
// overridden in memory mode).
func StartDisk() *StartCli {
	return StartDiskWithArgs([]string{})
}

// StartDiskWithArgs starts the node using disk-based storage with additional CLI arguments.
func StartDiskWithArgs(args []string) *StartCli {
	return &StartCli{
		inlineArgs: args,
		useDisk:    true,
	}
}

func (a *StartCli) Execute() {
	a.s.RootDir = a.s.T.TempDir()
	// Wazero is the only lens runtime supported on all platforms. Set it via
	// environment variable since there is no dedicated CLI flag for the lens runtime.
	// t.Setenv restores the original value when the test ends.
	a.s.T.Setenv("DEFRA_LENS_RUNTIME", "wazero")
	args := []string{
		"start",
		"--no-keyring",
		"--rootdir", a.s.RootDir,
		"--url=127.0.0.1:",
		"--document-acp-type=local",
	}
	if !a.useDisk {
		args = append(args, "--store=memory")
	}

	args = append(args, a.inlineArgs...)

	if a.expectedError != nil {
		select {
		case err := <-asyncExecute(a.s.Ctx, args):
			require.Contains(a.s.T, err.Error(), a.expectedError.Error())
		case <-time.After(1 * time.Second):
			a.s.T.Error("expected error but got none")
		}
		return
	}

	// Define a messageChans and store it on the context so that
	// we can be notified when the node is ready.
	messageChans := &node.MessageChans{
		APIURL: make(chan string),
	}
	ctx := node.SetContextMessageChans(a.s.Ctx, messageChans)

	select {
	case err := <-asyncExecute(ctx, args):
		require.NoError(a.s.T, err)
	case a.s.Url = <-messageChans.APIURL:
	case <-ctx.Done():
		a.s.T.Error("expected url but got none")
	}
}
