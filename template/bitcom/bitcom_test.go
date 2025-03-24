package bitcom

import (
	"bytes"
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	// Test nil script
	var nilScript *script.Script
	result := Decode(nilScript)
	require.NotNil(t, result, "Expected non-nil result for nil script")
	require.Empty(t, result.Protocols, "Expected empty protocols for nil script")
}

func TestLock(t *testing.T) {
	tests := []struct {
		name     string
		bitcom   *Bitcom
		expected []byte
	}{
		{
			name: "empty bitcom",
			bitcom: &Bitcom{
				ScriptPrefix: []byte{},
				Protocols:    []*BitcomProtocol{},
			},
			expected: []byte{},
		},
		{
			name: "bitcom with prefix only",
			bitcom: &Bitcom{
				ScriptPrefix: []byte{0x51}, // OP_1
				Protocols:    []*BitcomProtocol{},
			},
			expected: []byte{0x51},
		},
		{
			name: "bitcom with one protocol",
			bitcom: &Bitcom{
				ScriptPrefix: []byte{0x00}, // OP_FALSE
				Protocols: []*BitcomProtocol{
					{
						Protocol: MapPrefix,
						Script:   []byte("test data"),
					},
				},
			},
			expected: func() []byte {
				s := &script.Script{}
				s.AppendOpcodes(script.OpFALSE)
				s.AppendOpcodes(script.OpRETURN)
				s.AppendPushData([]byte(MapPrefix))
				s.AppendPushData([]byte("test data"))
				return *s
			}(),
		},
		{
			name: "bitcom with multiple protocols",
			bitcom: &Bitcom{
				ScriptPrefix: []byte{0x00}, // OP_FALSE
				Protocols: []*BitcomProtocol{
					{
						Protocol: MapPrefix,
						Script:   []byte("map data"),
					},
					{
						Protocol: BPrefix,
						Script:   []byte("b data"),
					},
				},
			},
			expected: func() []byte {
				s := &script.Script{}
				s.AppendOpcodes(script.OpFALSE)
				s.AppendOpcodes(script.OpRETURN)
				s.AppendPushData([]byte(MapPrefix))
				s.AppendPushData([]byte("map data"))
				s.AppendPushData([]byte("|"))
				s.AppendPushData([]byte(BPrefix))
				s.AppendPushData([]byte("b data"))
				return *s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bitcom.Lock()
			if len(tt.expected) == 0 {
				require.Equal(t, len(tt.expected), len(*result))
				return
			}
			require.True(t, bytes.Equal(tt.expected, *result), "expected %x but got %x", tt.expected, *result)
		})
	}
}

func TestFindReturn(t *testing.T) {
	// Test nil script
	var nilScript *script.Script
	result := findReturn(nilScript, 0)
	require.Equal(t, -1, result, "Expected -1 for nil script in findReturn")
}

func TestFindPipe(t *testing.T) {
	// Test nil script
	var nilScript *script.Script
	result := findPipe(nilScript, 0)
	require.Equal(t, -1, result, "Expected -1 for nil script in findPipe")
} 