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

//go:build javaclient

package tests

import "testing"

// TestPackageLinks exists purely to force `go test` to build (and link) this package, which
// imports tests/clients/java. Without any test file, `go test ./tests/clients/java/...` reports
// "no test files" and never attempts a link, silently hiding whether the java package's own
// dependency graph (see its link.go) actually resolves the Java_source_defra_* symbols
// nativewrapper.c references.
func TestPackageLinks(t *testing.T) {}
