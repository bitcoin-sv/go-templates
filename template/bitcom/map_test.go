package bitcom

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

func TestDecodeMap(t *testing.T) {
	// Test empty script
	emptyScript := script.Script{}
	result := DecodeMap(emptyScript)
	require.Nil(t, result, "Expected nil result for empty script")
	
	// Test DecodeMapBytes with nil bytes
	result = DecodeMapBytes(nil)
	require.Nil(t, result, "Expected nil result for nil bytes")
} 