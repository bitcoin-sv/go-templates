package ordp2pkh

import (
	"testing"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bitcoin-sv/go-templates/template/inscription"
	"github.com/bitcoin-sv/go-templates/template/p2pkh"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/require"
)

// TestOrdP2PKHDecode verifies that the Decode function properly identifies scripts
// that contain both an inscription and a P2PKH locking script
func TestOrdP2PKHDecode(t *testing.T) {
	// Create a private key and address
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)
	address, err := script.NewAddressFromPublicKey(privKey.PubKey(), true)
	require.NoError(t, err)

	// Create a P2PKH locking script
	p2pkhScript, err := p2pkh.Lock(address)
	require.NoError(t, err)
	require.NotNil(t, p2pkhScript)

	// Create a basic inscription
	inscr := &inscription.Inscription{
		File: inscription.File{
			Type:    "text/plain",
			Content: []byte("Hello, OrdP2PKH!"),
		},
		ScriptSuffix: *p2pkhScript,
	}

	// Create a combined script
	combinedScript, err := inscr.Lock()
	require.NoError(t, err)
	require.NotNil(t, combinedScript)

	// Decode the combined script
	decoded := Decode(combinedScript)
	require.NotNil(t, decoded, "Failed to decode OrdP2PKH script")

	// Verify that the decoded inscription data is correct
	require.NotNil(t, decoded.Inscription, "Inscription part is missing")

	// Check that the content is in Content field now
	require.Equal(t, "text/plain", decoded.Inscription.File.Type)
	require.Equal(t, "Hello, OrdP2PKH!", string(decoded.Inscription.File.Content))

	// Verify that the decoded P2PKH address is correct
	require.NotNil(t, decoded.Address, "Address part is missing")
	require.Equal(t, address.AddressString, decoded.Address.AddressString)
}

// TestOrdP2PKHDecodeInvalid verifies that the Decode function returns nil
// for scripts that don't match the OrdP2PKH pattern
func TestOrdP2PKHDecodeInvalid(t *testing.T) {
	// Test with a script that has inscription but no P2PKH
	inscr := &inscription.Inscription{
		File: inscription.File{
			Type:    "text/plain",
			Content: []byte("Hello, Ord!"),
		},
	}

	inscrScript, err := inscr.Lock()
	require.NoError(t, err)
	require.NotNil(t, inscrScript)

	// Try to decode - should be nil because no P2PKH
	decoded := Decode(inscrScript)
	require.Nil(t, decoded, "Should not decode script with inscription but no P2PKH")

	// Test with a script that has P2PKH but no inscription
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)
	address, err := script.NewAddressFromPublicKey(privKey.PubKey(), true)
	require.NoError(t, err)

	p2pkhScript, err := p2pkh.Lock(address)
	require.NoError(t, err)
	require.NotNil(t, p2pkhScript)

	// Try to decode - should be nil because no inscription
	decoded = Decode(p2pkhScript)
	require.Nil(t, decoded, "Should not decode script with P2PKH but no inscription")
}

// TestOrdP2PKHStructFields verifies that the OrdP2PKH struct fields work as expected
func TestOrdP2PKHStructFields(t *testing.T) {
	// Create an OrdP2PKH instance directly
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)
	address, err := script.NewAddressFromPublicKey(privKey.PubKey(), true)
	require.NoError(t, err)

	ordP2PKH := &OrdP2PKH{
		Inscription: &inscription.Inscription{
			File: inscription.File{
				Type:    "text/plain",
				Content: []byte("Hello, OrdP2PKH!"),
			},
		},
		Address: address,
	}

	// Verify that the fields are set correctly
	require.NotNil(t, ordP2PKH.Inscription)
	require.Equal(t, "text/plain", ordP2PKH.Inscription.File.Type)
	require.Equal(t, []byte("Hello, OrdP2PKH!"), ordP2PKH.Inscription.File.Content)
	require.Equal(t, address.AddressString, ordP2PKH.Address.AddressString)
}

// TestOrdP2PKHEndToEnd runs an end-to-end test of creating an OrdP2PKH script,
// decoding it, and verifying the result
func TestOrdP2PKHEndToEnd(t *testing.T) {
	// Create a private key and address
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)
	address, err := script.NewAddressFromPublicKey(privKey.PubKey(), true)
	require.NoError(t, err)

	// Create an OrdP2PKH instance with inscription and address
	ordP2PKH := &OrdP2PKH{
		Inscription: &inscription.Inscription{
			File: inscription.File{
				Type:    "image/png",
				Content: []byte("Simulated image data"),
			},
		},
		Address: address,
	}

	// Use the Lock method to create the combined script
	combinedScript, err := ordP2PKH.Lock()
	require.NoError(t, err)
	require.NotNil(t, combinedScript)

	// Log the script for debugging
	t.Logf("Combined script length: %d bytes", len(*combinedScript))

	// Decode the combined script
	decoded := Decode(combinedScript)
	require.NotNil(t, decoded)

	// Verify that we can extract both parts
	require.NotNil(t, decoded.Inscription)
	require.NotNil(t, decoded.Address)

	// Check that the content is in Content field now
	require.Equal(t, "image/png", decoded.Inscription.File.Type)
	require.Equal(t, "Simulated image data", string(decoded.Inscription.File.Content))

	// Check P2PKH address
	require.Equal(t, address.AddressString, decoded.Address.AddressString)
}

// TestOrdP2PKHWithMapMetadata tests creating an OrdP2PKH script with MAP metadata
func TestOrdP2PKHWithMapMetadata(t *testing.T) {
	// Create a private key and address
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)
	address, err := script.NewAddressFromPublicKey(privKey.PubKey(), true)
	require.NoError(t, err)

	// Create a P2PKH script
	p2pkhScript, err := p2pkh.Lock(address)
	require.NoError(t, err)

	// Create an inscription directly
	inscr := &inscription.Inscription{
		File: inscription.File{
			Type:    "image/png",
			Content: []byte("NFT image data"),
		},
		ScriptSuffix: *p2pkhScript,
	}

	// Lock the inscription directly to verify the format
	inscrScript, err := inscr.Lock()
	require.NoError(t, err)

	// Log the inscription script for debugging
	t.Logf("Pure inscription script length: %d bytes", len(*inscrScript))
	chunks, err := inscrScript.Chunks()
	require.NoError(t, err)

	// Check for 'ord' in raw inscription
	var foundOrdInRawInscr bool
	for i, chunk := range chunks {
		if len(chunk.Data) >= 3 && string(chunk.Data[:3]) == "ord" {
			foundOrdInRawInscr = true
			t.Logf("Found 'ord' marker in raw inscription at chunk index %d", i)
			break
		}
	}
	require.True(t, foundOrdInRawInscr, "No 'ord' inscription marker found in raw inscription")

	// Create an OrdP2PKH instance with inscription and address
	ordP2PKH := &OrdP2PKH{
		Inscription: &inscription.Inscription{
			File: inscription.File{
				Type:    "image/png",
				Content: []byte("NFT image data"),
			},
		},
		Address: address,
	}

	// Generate an OrdP2PKH script without MAP metadata first
	basicScript, err := ordP2PKH.Lock()
	require.NoError(t, err)

	// Check for 'ord' in basic OrdP2PKH script
	chunks, err = basicScript.Chunks()
	require.NoError(t, err)

	var foundOrdInBasic bool
	for i, chunk := range chunks {
		if len(chunk.Data) >= 3 && string(chunk.Data[:3]) == "ord" {
			foundOrdInBasic = true
			t.Logf("Found 'ord' marker in basic OrdP2PKH at chunk index %d", i)
			break
		}
	}
	require.True(t, foundOrdInBasic, "No 'ord' inscription marker found in basic OrdP2PKH")

	// Decode the basic script to verify it works
	decoded := Decode(basicScript)
	require.NotNil(t, decoded, "Failed to decode the basic script")
	require.NotNil(t, decoded.Inscription, "No inscription in decoded result")
	require.NotNil(t, decoded.Address, "No address in decoded result")
	require.Equal(t, address.AddressString, decoded.Address.AddressString, "Address doesn't match")

	// Check that the content is in Content field now
	require.Equal(t, "image/png", decoded.Inscription.File.Type)
	require.Equal(t, "NFT image data", string(decoded.Inscription.File.Content))

	// Create MAP metadata for an NFT
	metadata := &bitcom.Map{
		Cmd: bitcom.MapCmdSet,
		Data: map[string]string{
			"app":         "test-nft-app",
			"type":        "nft",
			"name":        "Test NFT",
			"description": "This is a test NFT with MAP metadata",
			"creator":     "Test User",
			"category":    "test",
		},
	}

	// Use the LockWithMapMetadata method to create the combined script
	combinedScript, err := ordP2PKH.LockWithMapMetadata(metadata)
	require.NoError(t, err)
	require.NotNil(t, combinedScript)

	// Log the combined script for debugging
	t.Logf("Combined script length: %d bytes", len(*combinedScript))
	chunks, err = combinedScript.Chunks()
	require.NoError(t, err)

	// Log all chunks for debugging
	t.Logf("Total chunks in combined script: %d", len(chunks))
	for i, chunk := range chunks {
		if chunk.Op == script.OpRETURN {
			t.Logf("Found OP_RETURN at chunk index %d", i)
			if i+1 < len(chunks) {
				t.Logf("  Next chunk data: %s", string(chunks[i+1].Data))
			}
		}

		// Check for 'ord' in any data chunk
		if len(chunk.Data) >= 3 {
			t.Logf("Chunk %d data prefix: %s", i, string(chunk.Data[:3]))
		}
	}

	// Check for the 'ord' inscription marker
	var foundOrdMarker bool
	for i, chunk := range chunks {
		if len(chunk.Data) >= 3 && string(chunk.Data[:3]) == "ord" {
			foundOrdMarker = true
			t.Logf("Found 'ord' marker at chunk index %d", i)
			break
		}
	}
	require.True(t, foundOrdMarker, "No 'ord' inscription marker found in the script")

	// When MAP data is appended, we only validate:
	// 1. That the 'ord' marker is still present in the combined script
	// 2. That when we use bitcom to decode the script, we can find the MAP protocol
	bc := bitcom.Decode(combinedScript)
	require.NotNil(t, bc, "Failed to decode the combined script with bitcom.Decode")

	// Look for MAP data in the decoded protocols
	var foundMAP bool
	for _, proto := range bc.Protocols {
		if proto.Protocol == bitcom.MapPrefix {
			foundMAP = true
			mapData := bitcom.DecodeMap(proto.Script)
			require.NotNil(t, mapData, "MAP data is nil")
			require.Equal(t, "SET", string(mapData.Cmd), "Expected MAP command to be 'SET'")
			require.Equal(t, metadata.Data["app"], mapData.Data["app"], "App field in MAP data doesn't match")
			require.Equal(t, metadata.Data["type"], mapData.Data["type"], "Type field in MAP data doesn't match")
			require.Equal(t, metadata.Data["name"], mapData.Data["name"], "Name field in MAP data doesn't match")
			require.Equal(t, metadata.Data["description"], mapData.Data["description"], "Description field in MAP data doesn't match")
			require.Equal(t, metadata.Data["creator"], mapData.Data["creator"], "Creator field in MAP data doesn't match")
			require.Equal(t, metadata.Data["category"], mapData.Data["category"], "Category field in MAP data doesn't match")
			break
		}
	}
	require.True(t, foundMAP, "MAP protocol not found in decoded protocols")
}

// TestLockWithAddress tests the convenience method for creating an OrdP2PKH script
func TestLockWithAddress(t *testing.T) {
	// Create a new private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Get the corresponding public key and address
	pubKey := privKey.PubKey()
	address, err := script.NewAddressFromPublicKey(pubKey, true)
	require.NoError(t, err)

	// Create an inscription
	inscription := &inscription.Inscription{
		File: inscription.File{
			Type:    "text/plain",
			Content: []byte("Convenience method test"),
		},
	}

	// Lock the inscription directly first to verify it works
	p2pkhScript, err := p2pkh.Lock(address)
	require.NoError(t, err)
	inscription.ScriptSuffix = *p2pkhScript

	directScript, err := inscription.Lock()
	require.NoError(t, err)
	require.NotNil(t, directScript)

	// Log the direct script chunks
	chunks, err := directScript.Chunks()
	require.NoError(t, err)

	// Check for 'ord' in the direct script
	var hasOrdDirect bool
	for i, chunk := range chunks {
		if len(chunk.Data) >= 3 && string(chunk.Data[:3]) == "ord" {
			hasOrdDirect = true
			t.Logf("Found 'ord' marker in direct script at chunk %d", i)
			break
		}
	}
	require.True(t, hasOrdDirect, "ord marker not found in direct script")

	// Create MAP metadata
	metadata := &bitcom.Map{
		Cmd: bitcom.MapCmdSet,
		Data: map[string]string{
			"app":        "test-app",
			"type":       "test-type",
			"test_field": "test_value",
		},
	}

	// Use the LockWithAddress convenience method
	combinedScript, err := LockWithAddress(address, inscription, metadata)
	require.NoError(t, err)
	require.NotNil(t, combinedScript)

	// Verify the combined script
	chunks, err = combinedScript.Chunks()
	require.NoError(t, err)

	// Debug log all chunks
	t.Logf("Combined script has %d chunks", len(chunks))
	for i, chunk := range chunks {
		if chunk.Op == script.OpRETURN {
			t.Logf("Found OP_RETURN at chunk %d", i)
			if i+1 < len(chunks) {
				t.Logf("  Next chunk data: %s", string(chunks[i+1].Data))
			}
		}

		if len(chunk.Data) >= 3 {
			t.Logf("Chunk %d data prefix: %s", i, string(chunk.Data[:3]))
		}
	}

	// Check for 'ord' marker
	hasOrd := false
	for i, chunk := range chunks {
		if len(chunk.Data) >= 3 && string(chunk.Data[:3]) == "ord" {
			hasOrd = true
			t.Logf("Found 'ord' marker in combined script at chunk %d", i)
			break
		}
	}
	require.True(t, hasOrd, "ord marker not found in script")

	// When MAP data is appended, we only validate:
	// 1. That the 'ord' marker is still present in the combined script
	// 2. That when we use bitcom to decode the script, we can find the MAP protocol
	bc := bitcom.Decode(combinedScript)
	require.NotNil(t, bc, "Failed to decode the combined script with bitcom.Decode")

	// Look for MAP data in the decoded protocols
	var foundMAP bool
	for _, proto := range bc.Protocols {
		if proto.Protocol == bitcom.MapPrefix {
			foundMAP = true
			mapData := bitcom.DecodeMap(proto.Script)
			require.NotNil(t, mapData, "MAP data is nil")
			require.Equal(t, "SET", string(mapData.Cmd), "Expected MAP command to be 'SET'")
			require.Equal(t, metadata.Data["app"], mapData.Data["app"], "App field in MAP data doesn't match")
			require.Equal(t, metadata.Data["type"], mapData.Data["type"], "Type field in MAP data doesn't match")
			require.Equal(t, metadata.Data["test_field"], mapData.Data["test_field"], "Test field in MAP data doesn't match")
			break
		}
	}
	require.True(t, foundMAP, "MAP protocol not found in decoded protocols")
}
