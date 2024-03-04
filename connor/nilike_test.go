package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNILike(t *testing.T) {
	const testString = "Source Is The Glue of Web3"

	// case insensitive exact match
	result, err := nilike("source is the glue of web3", testString)
	require.NoError(t, err)
	require.False(t, result)

	// case insensitive no match
	result, err = nilike("source is the glue", testString)
	require.NoError(t, err)
	require.True(t, result)

	// case insensitive match prefix
	result, err = nilike("source%", testString)
	require.NoError(t, err)
	require.False(t, result)

	// case insensitive match suffix
	result, err = nilike("%web3", testString)
	require.NoError(t, err)
	require.False(t, result)

	// case insensitive match contains
	result, err = nilike("%glue%", testString)
	require.NoError(t, err)
	require.False(t, result)

	// case insensitive match start and end with
	result, err = nilike("source%web3", testString)
	require.NoError(t, err)
	require.False(t, result)
}
