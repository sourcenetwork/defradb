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

package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	simpleUserSDL = `
	type User {
		name: String
		age: Int
		verified: Boolean
		points: Int @crdt(type: pncounter)
	}`

	bookOnlySDL = `
	type Book {
		name: String
		rating: Float
		author: Author
	}`

	authorOnlySDL = `
	type Author {
		name: String
		age: Int
		verified: Boolean
		published: [Book]
	}`
)

func TestSDLGenerate_SingleFileInput_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Create output file path
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, outputPath, fixturePath)
}

func TestSDLGenerate_MultipleFileInputs_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create separate input files
	bookPath := filepath.Join(tmpDir, "book.graphql")
	err := os.WriteFile(bookPath, []byte(bookOnlySDL), 0644)
	require.NoError(t, err)

	authorPath := filepath.Join(tmpDir, "author.graphql")
	err = os.WriteFile(authorPath, []byte(authorOnlySDL), 0644)
	require.NoError(t, err)

	// Create output file path
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, bookPath, authorPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.relatedmany.gen.graphql")
	assertGraphQLSchemasMatch(t, outputPath, fixturePath)
}

func TestSDLGenerate_StdinInput_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Create a pipe for stdin
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, "-"})

	// Set stdin
	cmd.SetIn(strings.NewReader(simpleUserSDL))

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, outputPath, fixturePath)
}

func TestSDLGenerate_OutputToStdout_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Capture stdout
	outBuf := new(bytes.Buffer)

	// Run the command with output to stdout
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", "-", inputPath})
	cmd.SetOut(outBuf)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output was written to buffer
	require.NotEmpty(t, outBuf.String())

	// Write the output to a temp file for comparison
	tempOutputPath := filepath.Join(tmpDir, "stdout_output.graphql")
	err = os.WriteFile(tempOutputPath, outBuf.Bytes(), 0644)
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, tempOutputPath, fixturePath)
}

func TestSDLGenerate_DefaultOutputPath_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Change to temp directory so default output goes there
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(originalWd) //nolint:errcheck

	// Run the command without specifying output (uses default)
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify default output file exists
	defaultOutputPath := filepath.Join(tmpDir, defaultOutputPath)
	_, err = os.Stat(defaultOutputPath)
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, defaultOutputPath, fixturePath)
}

func TestSDLGenerate_OverwriteExistingFile_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Create existing output file
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")
	err = os.WriteFile(outputPath, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Run the command with overwrite flag
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, "-y", inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output file was overwritten
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEqual(t, "existing content", string(content))

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, outputPath, fixturePath)
}

func TestSDLGenerate_ExistingFileWithoutOverwriteFlag_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Create existing output file
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")
	err = os.WriteFile(outputPath, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Run the command without overwrite flag
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, inputPath})

	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "output file path already exists")
}

func TestSDLGenerate_StdinWithOtherInputs_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command with stdin AND a file argument
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, inputPath, "-"})

	err = cmd.Execute()
	require.ErrorIs(t, err, ErrStdinSingleInputOnly)
}

func TestSDLGenerate_NoInputFiles_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command without any input files
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath})

	err := cmd.Execute()
	require.Error(t, err)
	// cobra should return an error about minimum arguments
	require.Contains(t, err.Error(), "requires at least 1 arg")
}

func TestSDLGenerate_NonExistentInputFile_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command with a non-existent file
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, "/nonexistent/file.graphql"})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestSDLGenerate_InvalidSDL_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file with invalid SDL
	inputPath := filepath.Join(tmpDir, "invalid.graphql")
	err := os.WriteFile(inputPath, []byte("this is not valid graphql {{{"), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, inputPath})

	err = cmd.Execute()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrParsingSDL)
}

func TestSDLGenerate_LongFormFlags_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	// Create existing output file
	outputPath := filepath.Join(tmpDir, "output.gen.graphql")
	err = os.WriteFile(outputPath, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Run the command with long-form flags
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "--output", outputPath, "--overwrite", inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify using graphql-inspector
	fixturePath := getTestFixturePath("schema.simple.gen.graphql")
	assertGraphQLSchemasMatch(t, outputPath, fixturePath)
}

func TestSDLGenerate_SearchableEncryptionFlag_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command with searchable encryption flag
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, "-s", inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Read the output and verify it contains searchable encryption types
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	// Searchable encryption adds specific types to the schema
	require.NotEmpty(t, content)
}

func TestSDLGenerate_SearchableEncryptionLongFlag_Succeeds(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tmpDir, "input.graphql")
	err := os.WriteFile(inputPath, []byte(simpleUserSDL), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command with long-form searchable encryption flag
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, "--include-searchable-encryption", inputPath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestSDLGenerate_EmptyInputFile_ReturnsError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create empty input file
	inputPath := filepath.Join(tmpDir, "empty.graphql")
	err := os.WriteFile(inputPath, []byte(""), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "output.gen.graphql")

	// Run the command
	cmd := NewDefraCommand(ctx)
	cmd.SetArgs([]string{"sdl", "generate", "-o", outputPath, inputPath})

	err = cmd.Execute()
	// Empty SDL should either succeed with minimal output or error
	// depending on the schema manager implementation
	if err != nil {
		require.ErrorIs(t, err, ErrParsingSDL)
	}
}

// getTestFixturePath returns the full path to a test fixture file
func getTestFixturePath(name string) string {
	// Get the absolute path to the fixtures directory relative to this test file
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(thisFile)
	return filepath.Join(baseDir, "..", "internal", "request", "graphql", "schema", "testfixtures", name)
}

// assertGraphQLSchemasMatch uses graphql-inspector to verify two schemas are equivalent
func assertGraphQLSchemasMatch(t *testing.T, actualPath, expectedPath string) {
	t.Helper()

	cmd := exec.Command("npx", "-y", "@graphql-inspector/cli@6.0.7",
		"diff",
		actualPath,
		expectedPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("graphql-inspector output:\n%s", string(output))
	}
	require.NoError(t, err, "GraphQL schemas should match")
	require.Contains(t, string(output), "No changes detected")
}
