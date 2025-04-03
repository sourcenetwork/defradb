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
}

var _ Action = (*StartCli)(nil)

func Start() *StartCli {
	return &StartCli{}
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

	logPrefix := "Providing GraphQL endpoint at "
	exampleUrl := "http://127.0.0.1:42571"

	logLine, err := executeUntil(a.s.Ctx, a.s, args, logPrefix)

	startIndex := strings.Index(logLine, logPrefix)
	// Take the url from the logs so that it may be passed into other commands later.
	// This is quite a lazy solution, if it breaks, strongly consider pulling it from the generated config
	// file instead.
	a.s.Url = logLine[startIndex+len(logPrefix) : startIndex+len(logPrefix)+len(exampleUrl)]

	require.NoError(a.s.T, err)
}
