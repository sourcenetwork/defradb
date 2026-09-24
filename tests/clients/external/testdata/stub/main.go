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

/*
Command stub stands in for the real defradb binary in external.Wrapper tests.
It accepts the same "start --url <host:port> ..." invocation Wrapper uses,
so wrapper_test.go can drive it without execing a real release binary.

Two modes, selected by the STUB_MODE environment variable:

  - "healthy" (default): serves the health-check endpoint on --url so a
    Wrapper can start successfully.
  - "unhealthy": writes a marker to stderr and blocks without ever serving
    the health-check endpoint, so a Wrapper's health wait/ctx times out.
*/
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Fprintln(os.Stderr, "stub: starting")

	var url string
	for i, arg := range os.Args {
		if arg == "--url" && i+1 < len(os.Args) {
			url = os.Args[i+1]
		}
	}

	if os.Getenv("STUB_MODE") == "unhealthy" {
		fmt.Fprintln(os.Stderr, "stub: unhealthy marker, never serving health-check")
		// Block forever without tripping the runtime's all-goroutines-asleep
		// deadlock detector; the parent test kills this process directly.
		<-time.After(24 * time.Hour)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("Healthy")
	})

	if err := http.ListenAndServe(url, mux); err != nil {
		fmt.Fprintln(os.Stderr, "stub: listen error:", err)
		os.Exit(1)
	}
}
