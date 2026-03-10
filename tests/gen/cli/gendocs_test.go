// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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

	ctx := context.Background()
	_, err := defra.db.AddCollection(ctx, `
	type User { 
		name: String 
		devices: [Device]
	}
	type Device {
		model: String
		owner: User
	}`)
	require.NoError(t, err)

	genDocsCmd := MakeGenDocCommand(ctx)
	outputBuf := bytes.NewBufferString("")
	genDocsCmd.SetOut(outputBuf)

	genDocsCmd.SetArgs([]string{
		"--demand", `{"User": 3, "Device": 12}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

	err = genDocsCmd.Execute()
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

	ctx := context.Background()
	_, err := defra.db.AddCollection(ctx, `
        type User { 
            name: String 
        }`)
	require.NoError(t, err)

	genDocsCmd := MakeGenDocCommand(ctx)
	genDocsCmd.SetArgs([]string{
		"--demand", `{"User": invalid}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

	err = genDocsCmd.Execute()
	require.ErrorContains(t, err, errInvalidDemandValue)
}

func TestGendocsCmd_IfInvalidConfig_ReturnError(t *testing.T) {
	defra, close := startTestNode(t)
	defer close()

	ctx := context.Background()
	_, err := defra.db.AddCollection(ctx, `
        type User { 
            name: String 
        }`)
	require.NoError(t, err)

	genDocsCmd := MakeGenDocCommand(ctx)
	genDocsCmd.SetArgs([]string{
		"--demand", `{"Unknown": 3}`,
		"--url", strings.TrimPrefix(defra.server.URL, "http://"),
	})

	err = genDocsCmd.Execute()
	require.Error(t, err, gen.NewErrInvalidConfiguration(""))
}
