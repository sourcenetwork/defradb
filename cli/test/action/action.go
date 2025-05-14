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
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sourcenetwork/testo/action"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/cli/test/state"
)

type Action = action.Action
type Actions = action.Actions
type Stateful = action.Stateful[*state.State]

// Augmented provides a function through which additional CLI arguments may be provided.
//
// It is typically used by multipliers in order to provide additional args to Actions
// implementing this interface.
type Augmented interface {
	// Augments the host object (typically an action) with additional CLI arguments.
	AddArgs(...string)
}

type AugmentedAction interface {
	Action
	Augmented
}

type stateful struct {
	s *state.State
}

var _ Stateful = (*stateful)(nil)

func (a *stateful) SetState(s *state.State) {
	if a == nil {
		a = &stateful{}
	}
	a.s = s
}

// AppendDirections appends the cli params responsible for letting the client know which
// Defra instance to act upon.
func (a *stateful) AppendDirections(args []string) []string {
	return append(args, "--url", a.s.Url, "--rootdir", a.s.RootDir)
}

type augmented struct {
	// Additional CLI arguments that should be sent along with the action-defined args.
	AdditionalArgs []string
}

var _ Augmented = (*augmented)(nil)

func (a *augmented) AddArgs(args ...string) {
	a.AdditionalArgs = append(a.AdditionalArgs, args...)
}

func executeJson[TReturn any](ctx context.Context, args []string) (TReturn, error) {
	var result TReturn

	data, err := executeBytes(ctx, args)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	return result, nil
}

func executeBytes(ctx context.Context, args []string) ([]byte, error) {
	stdOut, stdErr, err := executeStream(ctx, args)
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

func execute(ctx context.Context, args []string) error {
	_, err := executeBytes(ctx, args)
	return err
}

func executeSimple(ctx context.Context, args []string) error {
	cmd := cli.NewDefraCommand()
	cmd.SetArgs(args)

	return cmd.ExecuteContext(ctx)
}

func asyncExecute(ctx context.Context, args []string) chan error {
	err := make(chan error)
	go func() {
		err <- executeSimple(ctx, args)
		close(err)
	}()
	return err
}

func executeStream(ctx context.Context, args []string) (io.ReadCloser, io.ReadCloser, error) {
	stdOutRead, stdOutWrite := io.Pipe()
	stdErrRead, stdErrWrite := io.Pipe()

	cmd := cli.NewDefraCommand()
	cmd.SetOut(stdOutWrite)
	cmd.SetErr(stdErrWrite)
	cmd.SetArgs(args)

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	go func() {
		err := cmd.ExecuteContext(ctx)
		stdOutWrite.CloseWithError(err)
		stdErrWrite.CloseWithError(err)
	}()

	return stdOutRead, stdErrRead, nil
}
