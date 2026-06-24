// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
gen regenerates the committed SDL test fixtures from the current generator output.

It is an independent build step (run via `make sdl-fixtures`) and, unlike the
verification test, does not require node/npx.
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/testfixtures"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to generate SDL fixtures:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	for _, fixture := range testfixtures.Fixtures {
		sdl, err := testfixtures.GenerateSDL(ctx, fixture.SDL)
		if err != nil {
			return fmt.Errorf("generating %s: %w", fixture.Name, err)
		}

		path := testfixtures.Path(fixture.Name)
		if err := os.WriteFile(path, sdl, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}
