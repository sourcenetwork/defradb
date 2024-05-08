package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNLike(t *testing.T) {
	const testString = "Source is the glue of web3"

	// exact match
	result, err := nlike(testString, testString)
	require.NoError(t, err)
	require.False(t, result)

	// exact match error
	result, err = nlike("Source is the glue", testString)
	require.NoError(t, err)
	require.True(t, result)

	// match prefix
	result, err = nlike("Source%", testString)
	require.NoError(t, err)
	require.False(t, result)

	// match suffix
	result, err = nlike("%web3", testString)
	require.NoError(t, err)
	require.False(t, result)

	// match contains
	result, err = nlike("%glue%", testString)
	require.NoError(t, err)
	require.False(t, result)

	// match start and end with
	result, err = nlike("Source%web3", testString)
	require.NoError(t, err)
	require.False(t, result)
}
