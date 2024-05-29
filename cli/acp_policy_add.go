// Copyright 2024 Democratized Data Foundation
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
	"io"
	"os"

	"github.com/spf13/cobra"
)

func MakeACPPolicyAddCommand() *cobra.Command {
	const fileFlagLong string = "file"
	const fileFlagShort string = "f"

	var policyFile string

	var cmd = &cobra.Command{
		Use:   "add [-i --identity] [policy]",
		Short: "Add new policy",
		//nolint:lll
		Long: `Add new policy

Notes:
  - Can not add a policy without specifying an identity.
  - ACP must be available (i.e. ACP can not be disabled).
  - A non-DPI policy will be accepted (will be registered with acp system).
  - But only a valid DPI policyID & resource can be specified on a schema.
  - DPI validation happens when attempting to add a schema with '@policy'.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: add from an argument string:
  defradb client acp policy add -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f '
description: A Valid DefraDB Policy Interface

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
'

Example: add from file:
  defradb client acp policy add -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f -f policy.yml

Example: add from file, verbose flags:
  defradb client acp policy add --identity 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f --file policy.yml

Example: add from stdin:
  cat policy.yml | defradb client acp policy add -

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO-ACP: Ensure here (before going through acp system) if the required identity argument
			// is valid, if it is valid then keep proceeding further, otherwise return this error:
			// `NewErrRequiredFlagInvalid(identityFlagLongRequired, identityFlagShortRequired)`
			// Issue: https://github.com/sourcenetwork/defradb/issues/2358

			// Handle policy argument.
			extraArgsProvided := len(args)
			var policy string
			switch {
			case policyFile != "":
				data, err := os.ReadFile(policyFile)
				if err != nil {
					return err
				}
				policy = string(data)

			case extraArgsProvided > 0 && args[extraArgsProvided-1] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				policy = string(data)

			case extraArgsProvided > 0:
				policy = args[0]

			default:
				return ErrPolicyFileArgCanNotBeEmpty
			}

			db := mustGetContextDB(cmd)
			policyResult, err := db.AddPolicy(
				cmd.Context(),
				policy,
			)

			if err != nil {
				return err
			}

			return writeJSON(cmd, policyResult)
		},
	}
	cmd.Flags().StringVarP(&policyFile, fileFlagLong, fileFlagShort, "", "File to load a policy from")
	return cmd
}
