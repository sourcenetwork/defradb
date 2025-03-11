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
	"net"
	"strings"
	"time"

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
	// We need to lock across all processes in order to allocate the port.
	// We could alternatively let Defra assign the port and then read it from the logs,
	// but explicit assignment is currently the prefered approach.
	unlock, err := crossLock(44445)
	require.NoError(a.s.T, err)
	defer unlock()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(a.s.T, err)

	args := []string{"start", "--no-keyring", "--store=memory"}

	a.s.Url = listener.Addr().String()
	a.s.RootDir = a.s.T.TempDir()

	args = a.AppendDirections(args)

	err = listener.Close()
	require.NoError(a.s.T, err)

	err = executeUntil(a.s.Ctx, args, "Providing GraphQL endpoint at")
	require.NoError(a.s.T, err)
}

// crossLock forms a cross process lock by attempting to listen to the given port.
//
// This function will only return once the port is free or the timeout is reached.
// A function to unbind from the port is returned - this unlock function may be called
// multiple times without issue.
func crossLock(port uint16) (func(), error) {
	timeout := time.After(20 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout reached while trying to acquire cross process lock on port %v", port)
		default:
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", port))
			if err != nil {
				if strings.Contains(err.Error(), "address already in use") {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				return nil, err
			}

			return func() {
					// there are no errors that this returns that we actually care about
					_ = l.Close()
				},
				nil
		}
	}
}
