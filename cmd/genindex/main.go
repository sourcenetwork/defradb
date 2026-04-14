// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
genindex synchronises INDEX.md files under tests/integration with the current
state of *_test.go files.

For every Test* function referenced in an INDEX.md table row it updates:
  - the line range column to "startLine-endLine"
  - the description column to match the Description field in the TestCase struct

Usage:

	# sync every subdirectory under tests/integration (root is skipped)
	go run ./cmd/genindex

	# sync a single package (path relative to tests/integration)
	go run ./cmd/genindex --package subscription

	# sync a comma-separated list of packages
	go run ./cmd/genindex --package subscription,txn,node

	# sync using a glob (single-level wildcard)
	go run ./cmd/genindex --package "mutation/add/*"

	# sync using a recursive glob
	go run ./cmd/genindex --package "mutation/**"

	# combine globs and exact paths
	go run ./cmd/genindex --package "mutation/**,subscription,query/simple"
*/
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// testInfo holds the metadata extracted from a single Test* function.
type testInfo struct {
	startLine   int
	endLine     int
	description string
}

// rowRe matches a Markdown table row that references a Test* function.
//
// Expected format:
//
//	| `TestFunctionName` | 42 | Some description. |
//	| `TestFunctionName` | 42-99 | Some description. |
//
// Capture groups:
//
//	1 – function name
//	2 – current line / range value
//	3 – current description text
var rowRe = regexp.MustCompile(
	`^\| ` + "`" + `(Test\w+)` + "`" + ` \| (\d+(?:-\d+)?) \| (.+) \|$`,
)

func main() {
	pkg := flag.String(
		"package", "",
		"Comma-separated list of package paths or glob patterns relative to tests/integration.\n"+
			"Supports * (single path segment) and ** (any depth).\n"+
			"When empty, all subdirectories are processed.\n"+
			"Examples:\n"+
			"  subscription\n"+
			"  subscription,txn,node\n"+
			"  mutation/add/*\n"+
			"  mutation/**\n"+
			"  \"mutation/**,subscription\"",
	)
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		fatalf("finding repo root: %v", err)
	}

	integrationDir := filepath.Join(repoRoot, "tests", "integration")

	// Collect all subdirs once; we need them for glob expansion regardless.
	allDirs, err := collectSubdirs(integrationDir)
	if err != nil {
		fatalf("collecting subdirectories: %v", err)
	}

	var targets []string
	if *pkg == "" {
		targets = allDirs
	} else {
		targets, err = resolvePatterns(*pkg, integrationDir, allDirs)
		if err != nil {
			fatalf("%v", err)
		}
	}

	var errCount int
	for _, dir := range targets {
		if err := processDir(dir, integrationDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", dir, err)
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}

// resolvePatterns parses a comma-separated list of package path patterns and
// returns the matching absolute directory paths in the order they appear across
// allDirs (deduplication preserves walker order).
//
// Each pattern may be:
//   - an exact relative path:  "subscription"
//   - a single-level glob:     "mutation/add/*"
//   - a recursive glob:        "mutation/**"
//   - mixed with **:           "mutation/**/crdt"
func resolvePatterns(raw, integrationDir string, allDirs []string) ([]string, error) {
	patterns := strings.Split(raw, ",")
	for i := range patterns {
		patterns[i] = strings.TrimSpace(filepath.ToSlash(patterns[i]))
	}

	seen := make(map[string]bool)
	var result []string

	for _, dir := range allDirs {
		rel, err := filepath.Rel(integrationDir, dir)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)

		for _, pat := range patterns {
			if matchGlob(pat, rel) {
				if !seen[dir] {
					seen[dir] = true
					result = append(result, dir)
				}
				break
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no directories matched pattern(s): %s", raw)
	}
	return result, nil
}

// matchGlob reports whether path matches pattern using filepath.Match semantics
// for single-level wildcards, with additional support for ** which matches any
// number of path segments (including zero).
//
// Examples:
//
//	matchGlob("subscription",    "subscription")    → true
//	matchGlob("mutation/add/*",  "mutation/add/crdt") → true
//	matchGlob("mutation/add/*",  "mutation/add/field_kinds/one_to_many") → false
//	matchGlob("mutation/**",     "mutation/add/crdt") → true
//	matchGlob("mutation/**",     "mutation")         → true
//	matchGlob("**/crdt",         "mutation/add/crdt") → true
func matchGlob(pattern, path string) bool {
	// Fast path: no wildcards.
	if !strings.Contains(pattern, "*") {
		return pattern == path
	}

	// Split both into segments and use recursive matching.
	patSegs := strings.Split(pattern, "/")
	pathSegs := strings.Split(path, "/")
	return matchSegments(patSegs, pathSegs)
}

// matchSegments recursively matches pattern segments against path segments.
func matchSegments(pat, path []string) bool {
	if len(pat) == 0 {
		return len(path) == 0
	}

	if pat[0] == "**" {
		// ** matches zero or more segments: try consuming 0..len(path) segments.
		for skip := 0; skip <= len(path); skip++ {
			if matchSegments(pat[1:], path[skip:]) {
				return true
			}
		}
		return false
	}

	if len(path) == 0 {
		return false
	}

	ok, _ := filepath.Match(pat[0], path[0])
	if !ok {
		return false
	}
	return matchSegments(pat[1:], path[1:])
}

// findRepoRoot walks up from the working directory until it finds go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// collectSubdirs returns every subdirectory nested under integrationDir,
// excluding integrationDir itself (root-level tests are intentionally skipped).
func collectSubdirs(integrationDir string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(integrationDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != integrationDir {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs, err
}

// processDir parses the *_test.go files directly in dir (non-recursive),
// then updates dir/INDEX.md with fresh line ranges and descriptions.
func processDir(dir, integrationDir string) error {
	tests, err := parseTestsInDir(dir)
	if err != nil {
		return fmt.Errorf("parsing tests: %w", err)
	}
	if len(tests) == 0 {
		return nil // nothing to update
	}

	indexPath := filepath.Join(dir, "INDEX.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		rel, _ := filepath.Rel(integrationDir, dir)
		fmt.Printf("warning: no INDEX.md in %s (skipping)\n", rel)
		return nil
	}

	rel, _ := filepath.Rel(integrationDir, dir)
	fmt.Printf("syncing  %s\n", rel)
	return updateIndex(indexPath, tests)
}

// parseTestsInDir parses all *_test.go files directly in dir (not recursing
// into subdirectories) and returns a map from function name to testInfo.
func parseTestsInDir(dir string) (map[string]testInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	result := make(map[string]testInfo)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			result[fn.Name.Name] = testInfo{
				startLine:   fset.Position(fn.Pos()).Line,
				endLine:     fset.Position(fn.End()).Line,
				description: extractDescription(fn),
			}
		}
	}
	return result, nil
}

// extractDescription walks fn's body and returns the value of the first
// KeyValueExpr whose key is "Description" and whose value is a string literal.
// Returns "" if no such expression is found.
func extractDescription(fn *ast.FuncDecl) string {
	var desc string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if desc != "" {
			return false // already found
		}
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		ident, ok := kv.Key.(*ast.Ident)
		if !ok || ident.Name != "Description" {
			return true
		}
		lit, ok := kv.Value.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		s, err := strconv.Unquote(lit.Value)
		if err == nil {
			desc = s
		}
		return false
	})
	return desc
}

// updateIndex reads indexPath, rewrites every matching Test* table row with
// an updated line range and description, then writes the file back (only if
// something changed).
func updateIndex(indexPath string, tests map[string]testInfo) error {
	raw, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(raw), "\n")
	changed := false

	for i, line := range lines {
		m := rowRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		// m[1] = function name, m[2] = current range, m[3] = current description
		name := m[1]
		info, found := tests[name]
		if !found {
			fmt.Printf("  warning: %s in INDEX.md not found in test files\n", name)
			continue
		}

		lineRange := fmt.Sprintf("%d-%d", info.startLine, info.endLine)
		desc := m[3] // default: keep existing description
		if info.description != "" {
			desc = info.description
		}

		newLine := fmt.Sprintf("| `%s` | %s | %s |", name, lineRange, desc)
		if newLine != line {
			lines[i] = newLine
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return os.WriteFile(indexPath, []byte(strings.Join(lines, "\n")), 0o644)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "genindex: "+format+"\n", args...)
	os.Exit(1)
}
