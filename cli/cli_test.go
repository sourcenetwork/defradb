// Copyright 2025 Democratized Data Foundation
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
	"context"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/require"
)

func TestCLIExamples(t *testing.T) {
	ctx := context.Background()
	registry := newExampleRegistry()
	ctx = withExampleRegistry(ctx, registry)
	cmd := NewDefraCommand(ctx)

	for name, usage := range registry.examples {
		t.Run(name, func(t *testing.T) {
			args, err := shlex.Split(usage)
			require.NoError(t, err)
			err = validateCLIArgs(cmd, args[1:])
			require.NoError(t, err, "%s: `%s`", name, usage)
		})
	}
}
