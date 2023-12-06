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
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/tests/gen"
)

func execAddSchemaCmd(t *testing.T, cfg *config.Config, schema string) {
	rootCmd := cli.NewDefraCommand(cfg)
	rootCmd.SetArgs([]string{"client", "schema", "add", schema})
	err := rootCmd.Execute()
	require.NoError(t, err)
}

func TestGendocsCmd_IfNoErrors_ReturnGenerationOutput(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `
        type User { 
            name: String 
            devices: [Device]
        }
        type Device {
            model: String
            owner: User
        }`)

	genDocsCmd := MakeGenDocCommand(cfg)
	outputBuf := bytes.NewBufferString("")
	genDocsCmd.SetOut(outputBuf)

	genDocsCmd.SetArgs([]string{"--demand", `{"User": 3, "Device": 12}`})

	err := genDocsCmd.Execute()
	require.NoError(t, err)

	out, err := io.ReadAll(outputBuf)
	require.NoError(t, err)

	outStr := string(out)
	require.NoError(t, err)

	assert.Contains(t, outStr, "15")
	assert.Contains(t, outStr, "3")
	assert.Contains(t, outStr, "12")
	assert.Contains(t, outStr, "User")
	assert.Contains(t, outStr, "Device")
}

func TestGendocsCmd_IfInvalidDemandValue_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `
        type User { 
            name: String 
        }`)

	genDocsCmd := MakeGenDocCommand(cfg)
	genDocsCmd.SetArgs([]string{"--demand", `{"User": invalid}`})

	err := genDocsCmd.Execute()
	require.ErrorContains(t, err, errInvalidDemandValue)
}

func TestGendocsCmd_IfInvalidConfig_ReturnError(t *testing.T) {
	cfg, _, close := startTestNode(t)
	defer close()

	execAddSchemaCmd(t, cfg, `
        type User { 
            name: String 
        }`)

	genDocsCmd := MakeGenDocCommand(cfg)

	genDocsCmd.SetArgs([]string{"--demand", `{"Unknown": 3}`})

	err := genDocsCmd.Execute()
	require.Error(t, err, gen.NewErrInvalidConfiguration(""))
}
