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
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/gen"
)

func TestGendocsCmd_IfNoErrors_ReturnGenerationOutput(t *testing.T) {
	defra, close := startTestNode(t)
	defer close()

	defra.db.AddSchema(context.Background(), `
	type User { 
		name: String 
		devices: [Device]
	}
	type Device {
		model: String
		owner: User
	}`)

	genDocsCmd := MakeGenDocCommand(getTestConfig(t))
	outputBuf := bytes.NewBufferString("")
	genDocsCmd.SetOut(outputBuf)

	genDocsCmd.SetArgs([]string{
		"--demand", `{"User": 3, "Device": 12}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

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
	defra, close := startTestNode(t)
	defer close()

	defra.db.AddSchema(context.Background(), `
        type User { 
            name: String 
        }`)

	genDocsCmd := MakeGenDocCommand(getTestConfig(t))
	genDocsCmd.SetArgs([]string{
		"--demand", `{"User": invalid}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

	err := genDocsCmd.Execute()
	require.ErrorContains(t, err, errInvalidDemandValue)
}

func TestGendocsCmd_IfInvalidConfig_ReturnError(t *testing.T) {
	defra, close := startTestNode(t)
	defer close()

	defra.db.AddSchema(context.Background(), `
        type User { 
            name: String 
        }`)

	genDocsCmd := MakeGenDocCommand(getTestConfig(t))
	genDocsCmd.SetArgs([]string{
		"--demand", `{"Unknown": 3}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

	err := genDocsCmd.Execute()
	require.Error(t, err, gen.NewErrInvalidConfiguration(""))
}
