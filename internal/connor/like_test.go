package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLike(t *testing.T) {
	const testString = "Source is the glue of web3"

	// exact match
	result, err := like(testString, testString)
	require.NoError(t, err)
	require.True(t, result)

	// exact match error
	result, err = like("Source is the glue", testString)
	require.NoError(t, err)
	require.False(t, result)

	// match prefix
	result, err = like("Source%", testString)
	require.NoError(t, err)
	require.True(t, result)

	// match suffix
	result, err = like("%web3", testString)
	require.NoError(t, err)
	require.True(t, result)

	// match contains
	result, err = like("%glue%", testString)
	require.NoError(t, err)
	require.True(t, result)

	// match start and end with
	result, err = like("Source%web3", testString)
	require.NoError(t, err)
	require.True(t, result)
}
