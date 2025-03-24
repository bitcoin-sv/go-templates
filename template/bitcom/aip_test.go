package bitcom

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

func TestDecodeAIP(t *testing.T) {
	t.Run("nil bitcom", func(t *testing.T) {
		// Test nil Bitcom
		var nilBitcom *Bitcom
		result := DecodeAIP(nilBitcom)
		require.NotNil(t, result, "Result should be an empty slice, not nil")
		require.Empty(t, result, "Result should be an empty slice for nil Bitcom")
	})
	
	t.Run("empty protocols", func(t *testing.T) {
		// Test Bitcom with empty protocols
		emptyBitcom := &Bitcom{
			Protocols: []*BitcomProtocol{},
		}
		result := DecodeAIP(emptyBitcom)
		require.NotNil(t, result, "Result should be an empty slice, not nil")
		require.Empty(t, result, "Result should be an empty slice for Bitcom with empty protocols")
	})

	tests := []struct {
		name     string
		bitcom   *Bitcom
		expected []*AIP
	}{
		{
			name:     "protocols without AIP",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: MapPrefix,
						Script:   []byte("some data"),
					},
					{
						Protocol: BPrefix,
						Script:   []byte("more data"),
					},
				},
			},
			expected: []*AIP{},
		},
		{
			name: "valid AIP protocol with minimum fields",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: AIPPrefix,
						Script: func() []byte {
							s := &script.Script{}
							s.AppendPushData([]byte("BITCOIN_ECDSA"))
							s.AppendPushData([]byte("1address1234567890"))
							s.AppendPushData([]byte("signature1234567890"))
							return *s
						}(),
						Pos: 0,
					},
				},
			},
			expected: []*AIP{
				{
					Algorithm: "BITCOIN_ECDSA",
					Address:   "1address1234567890",
					Signature: "signature1234567890",
				},
			},
		},
		{
			name: "valid AIP protocol with field indexes",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: AIPPrefix,
						Script: func() []byte {
							s := &script.Script{}
							s.AppendPushData([]byte("BITCOIN_ECDSA"))
							s.AppendPushData([]byte("1address1234567890"))
							s.AppendPushData([]byte("signature1234567890"))
							s.AppendPushData([]byte("1"))
							s.AppendPushData([]byte("2"))
							s.AppendPushData([]byte("3"))
							return *s
						}(),
						Pos: 0,
					},
				},
			},
			expected: []*AIP{
				{
					Algorithm:    "BITCOIN_ECDSA",
					Address:      "1address1234567890",
					Signature:    "signature1234567890",
					FieldIndexes: []int{1, 2, 3},
				},
			},
		},
		{
			name: "multiple AIP protocols",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: AIPPrefix,
						Script: func() []byte {
							s := &script.Script{}
							s.AppendPushData([]byte("BITCOIN_ECDSA"))
							s.AppendPushData([]byte("1address1234567890"))
							s.AppendPushData([]byte("signature1234567890"))
							return *s
						}(),
						Pos: 0,
					},
				},
			},
			expected: []*AIP{
				{
					Algorithm: "BITCOIN_ECDSA",
					Address:   "1address1234567890",
					Signature: "signature1234567890",
				},
			},
		},
		{
			name: "invalid AIP protocol (missing fields)",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: AIPPrefix,
						Script: func() []byte {
							s := &script.Script{}
							s.AppendPushData([]byte("BITCOIN_ECDSA"))
							// Missing address and signature
							return *s
						}(),
						Pos: 0,
					},
				},
			},
			expected: []*AIP{},
		},
		{
			name: "invalid field indexes",
			bitcom: &Bitcom{
				Protocols: []*BitcomProtocol{
					{
						Protocol: AIPPrefix,
						Script: func() []byte {
							s := &script.Script{}
							s.AppendPushData([]byte("BITCOIN_ECDSA"))
							s.AppendPushData([]byte("1address1234567890"))
							s.AppendPushData([]byte("signature1234567890"))
							s.AppendPushData([]byte("not-a-number"))
							return *s
						}(),
						Pos: 0,
					},
				},
			},
			expected: []*AIP{
				{
					Algorithm:    "BITCOIN_ECDSA",
					Address:      "1address1234567890",
					Signature:    "signature1234567890",
					FieldIndexes: []int{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeAIP(tt.bitcom)
			
			// Debug output for multiple AIP protocols test
			if tt.name == "multiple AIP protocols" {
				t.Logf("Expected %d AIPs, got %d AIPs", len(tt.expected), len(result))
				t.Logf("Result: %+v", result)
			}
			
			require.Equal(t, len(tt.expected), len(result))
			
			if len(tt.expected) > 0 {
				for i, expectedAIP := range tt.expected {
					if i >= len(result) {
						t.Fatalf("Missing expected AIP at index %d", i)
						continue
					}
					
					resultAIP := result[i]
					require.Equal(t, expectedAIP.Algorithm, resultAIP.Algorithm)
					require.Equal(t, expectedAIP.Address, resultAIP.Address)
					require.Equal(t, expectedAIP.Signature, resultAIP.Signature)
					
					require.Equal(t, len(expectedAIP.FieldIndexes), len(resultAIP.FieldIndexes))
					for j, expectedIndex := range expectedAIP.FieldIndexes {
						if j >= len(resultAIP.FieldIndexes) {
							t.Fatalf("Missing expected field index at position %d", j)
							continue
						}
						require.Equal(t, expectedIndex, resultAIP.FieldIndexes[j])
					}
				}
			}
		})
	}
} 