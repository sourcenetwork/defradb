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
	"strings"

	"github.com/stretchr/testify/require"
)

type StartCli struct {
	stateful
	inlineArgs    []string
	expectedError error
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

func (a *StartCli) Execute() {
	a.s.RootDir = a.s.T.TempDir()
	args := []string{
		"start",
		"--no-keyring",
		"--store=memory",
		"--rootdir", a.s.RootDir,
		"--url=127.0.0.1:",
		"--acp-type=local",
	}

	args = append(args, a.inlineArgs...)

	logPrefix := "Providing GraphQL endpoint at "
	exampleUrl := "http://127.0.0.1:42571"

	// If we expect an error, then we will seek for it...
	if a.expectedError != nil {
		readLine, err := executeUntil(a.s.Ctx, a.s, args, a.expectedError.Error())
		require.NoError(a.s.T, err)
		require.Contains(a.s.T, readLine, a.expectedError.Error())
		return
	}

	// ...otherwise, we will seek for the logPrefix indicating that the service has started
	logLine, err := executeUntil(a.s.Ctx, a.s, args, logPrefix)
	startIndex := strings.Index(logLine, logPrefix)
	a.s.Url = logLine[startIndex+len(logPrefix) : startIndex+len(logPrefix)+len(exampleUrl)]

	require.NoError(a.s.T, err)
}
