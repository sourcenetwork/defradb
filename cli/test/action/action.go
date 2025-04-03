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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcenetwork/testo/action"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/cli/test/state"
	"github.com/sourcenetwork/defradb/errors"
)

var stdErr *os.File = os.Stderr

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

// executeUntil executes a defra command with the given args and will block until it reads the given
// `until` string in stderr.
func executeUntil(ctx context.Context, s *state.State, args []string, until string) (string, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return "", err
	}

	cmd := cli.NewDefraCommand()
	os.Stderr = write
	cmd.SetArgs(args)

	var wg sync.WaitGroup
	wg.Add(1)

	var targetLine string
	scanner := bufio.NewScanner(read)
	go func() {
		targetReached := false
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			hasValue := scanner.Scan()
			if !hasValue {
				time.Sleep(100 * time.Microsecond)
				continue
			}

			// The scanner splits by new line (useful), we have to remember to add it back in.
			// We could write our own split function to avoid this, but this is easier for now.
			_, err := stdErr.Write(append(scanner.Bytes(), []byte("\n")...))
			if err != nil {
				panic(err)
			}

			if !targetReached {
				targetReached = strings.Contains(scanner.Text(), until)
				if targetReached {
					targetLine = scanner.Text()
					wg.Add(-1)
				}
			}
		}
	}()

	var isDoneWg sync.WaitGroup
	isDoneWg.Add(1)
	go func() {
		err := cmd.ExecuteContext(ctx)
		if err != nil {
			_, innerErr := stdErr.WriteString(err.Error())
			if innerErr != nil {
				panic(errors.Join(err, innerErr))
			}
		}
		err = write.Close()
		if err != nil {
			panic(err)
		}
		isDoneWg.Done()
	}()

	originalWait := s.Wait
	s.Wait = func() {
		originalWait()
		// Because we are overwriting a singleton (os.Stderr) we must wait for the full command
		// to finish executing before we allow the next test to execute, else the next test may
		// try and overwrite os.Stderr before this execution finishes closing down the database.
		//
		// That might not cause any real bother, but the go test `-race` flag will complain.
		isDoneWg.Wait()
		os.Stderr = stdErr
	}

	// Block, until `until` is read from stderr
	wg.Wait()

	return targetLine, nil
}
