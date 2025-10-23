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

	// t.Fail()
}
