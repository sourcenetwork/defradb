// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build npx
// +build npx

package schema_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/testfixtures"
)

// TestWriteSDL verifies that each committed SDL fixture matches the current generator
// output. Regenerate the fixtures with `make sdl-fixtures` when the generator changes.
func TestWriteSDL(t *testing.T) {
	for _, fixture := range testfixtures.Fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			sdlResult, err := testfixtures.GenerateSDL(t.Context(), fixture.SDL)
			require.NoError(t, err)

			cmd := exec.Command("npx", "-y", "@graphql-inspector/cli@6.0.7",
				"diff",
				string(sdlResult),
				testfixtures.Path(fixture.Name))

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Log(string(output))
			}
			require.NoError(t, err)
			require.Contains(t, string(output), "No changes detected")
		})
	}
}
