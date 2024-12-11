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
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"
)

func MakeViewAddCommand() *cobra.Command {
	var queryFile string
	var sdlFile string
	var lensFile string
	cmd := &cobra.Command{
		Use:   "add [query] [sdl] [transform]",
		Short: "Add new view",
		Long: `Add new database view.

Example: add from an argument string:
  defradb client view add 'Foo { name, ...}' 'type Foo { ... }' '{"lenses": [...'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			fileOrArg := newFileOrArgData(args, os.ReadFile)
			query, err := fileOrArg.next(queryFile)
			if err != nil {
				return err
			}
			sdl, err := fileOrArg.next(sdlFile)
			if err != nil {
				return err
			}
			lensCfgJson, err := fileOrArg.next(lensFile)
			if err != nil {
				return err
			}

			if lensCfgJson == "-" {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			}

			var transform immutable.Option[model.Lens]
			if lensCfgJson != "" {
				decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
				decoder.DisallowUnknownFields()

				var lensCfg model.Lens
				if err := decoder.Decode(&lensCfg); err != nil {
					return NewErrInvalidLensConfig(err)
				}
				transform = immutable.Some(lensCfg)
			}

			defs, err := store.AddView(cmd.Context(), query, sdl, transform)
			if err != nil {
				return err
			}
			return writeJSON(cmd, defs)
		},
	}
	cmd.Flags().StringVarP(&lensFile, "file", "f", "", "Lens configuration file")
	cmd.Flags().StringVarP(&queryFile, "query-file", "", "", "Query file")
	cmd.Flags().StringVarP(&sdlFile, "sdl-file", "", "", "SDL file")
	return cmd
}

type readFileFn func(string) ([]byte, error)

// FileOrArgData tracks a serie of args.
type FileOrArgData struct {
	args         []string
	currentIndex int
	readFile     readFileFn
}

func newFileOrArgData(args []string, readFile readFileFn) FileOrArgData {
	return FileOrArgData{
		args:         args,
		currentIndex: 0,
		readFile:     readFile,
	}
}

// next gets the data primarily from a file when filePath is set, or from expected arg index.
func (x *FileOrArgData) next(filePath string) (string, error) {
	if filePath == "" {
		data := x.args[x.currentIndex]
		x.currentIndex += 1
		return data, nil
	}
	data, err := x.readFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
