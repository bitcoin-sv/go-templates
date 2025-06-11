package shrug

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

func TestLockAndDecode_RoundTrip(t *testing.T) {
	t.Skip("No valid test vector for shrug round-trip; skipping until one is available.")
	// TODO: add test vector
}

func TestDecode_InvalidScripts(t *testing.T) {
	// Not enough ops
	invalid := script.Script([]byte{0x00, 0x01})
	result := Decode(&invalid)
	require.Nil(t, result)

	// Wrong tag
	s := &script.Script{}
	_ = s.AppendPushData([]byte("notashrug"))
	result = Decode(s)
	require.Nil(t, result)
}
