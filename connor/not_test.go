package connor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNot_WithNotAndNotNot_NoError(t *testing.T) {
	const testString = "Source is the glue of web3"

	// not equal
	result, err := not(testString, testString)
	require.NoError(t, err)
	require.False(t, result)

	// not not equal
	result, err = not("Source is the glue", testString)
	require.NoError(t, err)
	require.True(t, result)
}

func TestNot_WithEmptyCondition_ReturnError(t *testing.T) {
	const testString = "Source is the glue of web3"

	_, err := not(map[FilterKey]any{&operator{"_some"}: "test"}, testString)
	require.ErrorIs(t, err, ErrUnknownOperator)
}

type operator struct {
	// The filter operation string that this `operator`` represents.
	//
	// E.g. "_eq", or "_and".
	Operation string
}

func (k *operator) GetProp(data any) any {
	return data
}

func (k *operator) GetOperatorOrDefault(defaultOp string) string {
	return k.Operation
}

func (k *operator) Equal(other FilterKey) bool {
	if otherKey, isOk := other.(*operator); isOk && *k == *otherKey {
		return true
	}
	return false
}
