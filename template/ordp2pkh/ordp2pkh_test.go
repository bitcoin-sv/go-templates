package ordp2pkh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bitcoin-sv/go-templates/template/inscription"
	"github.com/bitcoin-sv/go-templates/template/p2pkh"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
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

// TestDecodeRealOrdinalTransaction verifies that the OrdP2PKH can properly decode
// a real-world ordinal transaction
func TestDecodeRealOrdinalTransaction(t *testing.T) {
	// Transaction ID from the filename
	txID := "b08538c963d2b88c7d26600a1c3c925a3388e942cdc5f903ecf0009f18c41ff3"
	testdataFile := filepath.Join("testdata", txID+".hex")

	// Load the hex data from the file
	hexBytes, err := os.ReadFile(testdataFile)
	require.NoError(t, err, "Failed to read test vector file")

	// Create a transaction from the bytes
	tx, err := transaction.NewTransactionFromHex(strings.TrimSpace(string(hexBytes)))
	require.NoError(t, err, "Failed to parse transaction")

	// Verify transaction ID matches expected
	require.Equal(t, txID, tx.TxID().String(), "Transaction ID should match expected value")

	// Log transaction structure
	t.Logf("Transaction ID: %s", tx.TxID().String())
	t.Logf("Transaction has %d inputs and %d outputs", len(tx.Inputs), len(tx.Outputs))

	// Log satoshis for each output
	for i, output := range tx.Outputs {
		t.Logf("Output %d: %d satoshis", i, output.Satoshis)
	}

	// Try to find an output with an ordinal inscription
	var foundOrdP2PKH *OrdP2PKH
	var outputIndex int

	for i, output := range tx.Outputs {
		if output.LockingScript == nil || len(*output.LockingScript) == 0 {
			continue
		}

		decoded := Decode(output.LockingScript)
		if decoded != nil {
			foundOrdP2PKH = decoded
			outputIndex = i
			break
		}
	}

	// With our improved getAddressFromScript function, we should now find the OrdP2PKH
	require.NotNil(t, foundOrdP2PKH, "Should find an OrdP2PKH in output 0 with improved implementation")
	require.Equal(t, 0, outputIndex, "OrdP2PKH should be found in output 0")

	t.Logf("Found OrdP2PKH in output %d", outputIndex)

	// Log detailed info about the found ordinal
	t.Logf("Inscription Content Type: %s", foundOrdP2PKH.Inscription.File.Type)
	t.Logf("Inscription Content Size: %d bytes", len(foundOrdP2PKH.Inscription.File.Content))

	// Log address info if available
	require.NotNil(t, foundOrdP2PKH.Address, "Address should not be nil")
	t.Logf("P2PKH Address: %s", foundOrdP2PKH.Address.AddressString)

	// Verify the address is the expected one
	expectedAddress := "1Cr5gSHf5tzFBvGuSa21VRoV9pRuRBmum9"
	require.Equal(t, expectedAddress, foundOrdP2PKH.Address.AddressString, "Address should match expected value")

	// Log metadata if available
	if foundOrdP2PKH.Metadata != nil {
		t.Logf("MAP Metadata found with %d fields", len(foundOrdP2PKH.Metadata.Data))
		for key, value := range foundOrdP2PKH.Metadata.Data {
			t.Logf("  %s: %s", key, value)
		}
	}
}

// TestRobustP2PKHExtraction demonstrates the issue with the current implementation
// and proposes a fix for parsing OrdP2PKH from scripts that contain additional data
// after the P2PKH part
func TestRobustP2PKHExtraction(t *testing.T) {
	// Transaction ID from the filename
	txID := "b08538c963d2b88c7d26600a1c3c925a3388e942cdc5f903ecf0009f18c41ff3"
	testdataFile := filepath.Join("testdata", txID+".hex")

	// Load the hex data from the file
	hexBytes, err := os.ReadFile(testdataFile)
	require.NoError(t, err, "Failed to read test vector file")

	// Create a transaction from the bytes
	tx, err := transaction.NewTransactionFromHex(strings.TrimSpace(string(hexBytes)))
	require.NoError(t, err, "Failed to parse transaction")

	// Directly examine output 0 which contains the ordinal
	require.GreaterOrEqual(t, len(tx.Outputs), 1, "Transaction should have at least one output")
	output0 := tx.Outputs[0]
	require.NotNil(t, output0.LockingScript, "Output 0 should have a locking script")

	// Try to decode the inscription
	inscr := inscription.Decode(output0.LockingScript)
	require.NotNil(t, inscr, "Should be able to decode inscription in output 0")

	t.Logf("Found inscription in output 0:")
	t.Logf("  Content Type: %s", inscr.File.Type)
	t.Logf("  Content Size: %d bytes", len(inscr.File.Content))
	t.Logf("  Suffix Script: %d bytes", len(inscr.ScriptSuffix))

	// Standard getAddressFromScript function check
	standardAddr := getAddressFromScript(inscr)
	if standardAddr != nil {
		t.Logf("Standard getAddressFromScript found address: %s", standardAddr.AddressString)
	} else {
		t.Logf("Standard getAddressFromScript did not find an address")
	}

	// More robust implementation to handle scripts with additional data after P2PKH
	robustAddr := getAddressFromScriptRobust(inscr)
	require.NotNil(t, robustAddr, "Robust implementation should find a P2PKH address")
	t.Logf("Robust implementation found address: %s", robustAddr.AddressString)

	// Create a manual OrdP2PKH with the address found
	robustOrdP2PKH := &OrdP2PKH{
		Inscription: inscr,
		Address:     robustAddr,
	}

	// Test that our custom OrdP2PKH with the robust function works
	t.Logf("Created OrdP2PKH with address: %s", robustOrdP2PKH.Address.AddressString)
	t.Logf("Content Type: %s", robustOrdP2PKH.Inscription.File.Type)
}

// getAddressFromScriptRobust is a more robust version of getAddressFromScript that can
// extract a P2PKH address even when the script contains additional data after the P2PKH part
func getAddressFromScriptRobust(inscription *inscription.Inscription) *script.Address {
	// First try the standard method
	if addr := getAddressFromScript(inscription); addr != nil {
		return addr
	}

	// Check suffix for embedded P2PKH
	if len(inscription.ScriptSuffix) >= 25 {
		suffix := script.NewFromBytes(inscription.ScriptSuffix)
		chunks, err := suffix.Chunks()
		if err != nil {
			return nil
		}

		// Look for P2PKH pattern (OP_DUP OP_HASH160 <pubkeyhash> OP_EQUALVERIFY OP_CHECKSIG)
		// at the beginning of the script
		if len(chunks) >= 5 &&
			chunks[0].Op == script.OpDUP &&
			chunks[1].Op == script.OpHASH160 &&
			len(chunks[2].Data) == 20 &&
			chunks[3].Op == script.OpEQUALVERIFY &&
			chunks[4].Op == script.OpCHECKSIG {

			// Extract just the P2PKH part
			p2pkhPart := &script.Script{}
			_ = p2pkhPart.AppendOpcodes(script.OpDUP, script.OpHASH160)
			_ = p2pkhPart.AppendPushData(chunks[2].Data)
			_ = p2pkhPart.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIG)

			// Check if this is a valid P2PKH script
			return p2pkh.Decode(p2pkhPart, true)
		}
	}

	// Check prefix as well for completeness
	if len(inscription.ScriptPrefix) >= 25 {
		prefix := script.NewFromBytes(inscription.ScriptPrefix)
		chunks, err := prefix.Chunks()
		if err != nil {
			return nil
		}

		// Look for P2PKH pattern at the beginning of the script
		if len(chunks) >= 5 &&
			chunks[0].Op == script.OpDUP &&
			chunks[1].Op == script.OpHASH160 &&
			len(chunks[2].Data) == 20 &&
			chunks[3].Op == script.OpEQUALVERIFY &&
			chunks[4].Op == script.OpCHECKSIG {

			// Extract just the P2PKH part
			p2pkhPart := &script.Script{}
			_ = p2pkhPart.AppendOpcodes(script.OpDUP, script.OpHASH160)
			_ = p2pkhPart.AppendPushData(chunks[2].Data)
			_ = p2pkhPart.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIG)

			// Check if this is a valid P2PKH script
			return p2pkh.Decode(p2pkhPart, true)
		}
	}

	return nil
}
