package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestILike(t *testing.T) {
	const testString = "Source Is The Glue of Web3"

	// case insensitive exact match
	result, err := ilike("source is the glue of web3", testString)
	require.NoError(t, err)
	require.True(t, result)

	// case insensitive no match
	result, err = ilike("source is the glue", testString)
	require.NoError(t, err)
	require.False(t, result)

	// case insensitive match prefix
	result, err = ilike("source%", testString)
	require.NoError(t, err)
	require.True(t, result)

	// case insensitive match suffix
	result, err = ilike("%web3", testString)
	require.NoError(t, err)
	require.True(t, result)

	// case insensitive match contains
	result, err = ilike("%glue%", testString)
	require.NoError(t, err)
	require.True(t, result)

	// case insensitive match start and end with
	result, err = ilike("source%web3", testString)
	require.NoError(t, err)
	require.True(t, result)
}
