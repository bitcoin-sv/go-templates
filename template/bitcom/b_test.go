package bitcom

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

func TestDecodeB(t *testing.T) {
	// Test nil script
	var nilScript *script.Script
	result := DecodeB(nilScript)
	require.Nil(t, result, "Expected nil result for nil script")
	
	// Test DecodeBBytes with nil bytes
	result = DecodeBBytes(nil)
	require.Nil(t, result, "Expected nil result for nil bytes")
} 