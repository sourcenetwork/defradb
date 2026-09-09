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

package schema

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	"github.com/sourcenetwork/defradb/client"
)

func TestIntrospectionResult(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionVersion{
		{
			Name: "User",
			Fields: []client.CollectionFieldDescription{
				{Name: "email", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "ssn", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
			},
		},
	}

	err = manager.Generate(ctx, collections)
	require.NoError(t, err)

	var report operationreport.Report
	var data introspection.Data
	introspection.NewGenerator().Generate(manager.Definition(), &report, &data)
	require.False(t, report.HasErrors(), report.Error())

	tempDir := t.TempDir()
	resultFileName := filepath.Join(tempDir, "introspection_data2.json")
	filebuf, err := json.Marshal(data)
	require.NoError(t, err)
	err = os.WriteFile(resultFileName, filebuf, 0644)
	require.NoError(t, err)

	cmd := exec.Command("npx", "-y", "graphql-introspection-json-to-sdl", resultFileName)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	// this check is mostlyy redundent relative to the err check above, but im leaving it in all the same
	require.False(t, strings.HasPrefix(string(output), "Error:"))
}
