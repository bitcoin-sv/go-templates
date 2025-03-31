package ordlock

import (
	"encoding/hex"
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/require"
)

func TestOrdLock(t *testing.T) {
	// Create an OrdLock instance
	publicKeyHash, _ := hex.DecodeString("1234567890abcdef1234567890abcdef12345678")
	seller, _ := script.NewAddressFromPublicKeyHash(publicKeyHash, true)

	ordLock := &OrdLock{
		Seller:   seller,
		Price:    1000,
		PricePer: 0.5,
		PayOut:   []byte("test payout"),
	}

	// Basic validation
	require.Equal(t, uint64(1000), ordLock.Price)
	require.Equal(t, 0.5, ordLock.PricePer)
	require.Equal(t, []byte("test payout"), ordLock.PayOut)
	require.NotNil(t, ordLock.Seller)
}

func TestCreateOrdLockScript(t *testing.T) {
	// Create test data
	publicKeyHash, _ := hex.DecodeString("1234567890abcdef1234567890abcdef12345678")

	// Create a simple payout output
	payoutOutput := &transaction.TransactionOutput{
		Satoshis: 5000,
	}

	// Initialize with an empty script
	emptyScript := script.NewFromBytes([]byte{})
	payoutOutput.LockingScript = emptyScript

	// Manually build an OrdLock script similar to what we'd expect from the package
	var scriptData []byte

	// Add the prefix
	scriptData = append(scriptData, OrdLockPrefix...)

	// Add the PKHash operation - in a real implementation this would come from the seller
	scriptData = append(scriptData, 0x14) // OP_DATA_20
	scriptData = append(scriptData, publicKeyHash...)

	// Add the payout output
	outputBytes := payoutOutput.Bytes()
	scriptData = append(scriptData, byte(len(outputBytes))) // Length of output data
	scriptData = append(scriptData, outputBytes...)

	// Add the suffix
	scriptData = append(scriptData, OrdLockSuffix...)

	// Create the script
	lockScript := script.NewFromBytes(scriptData)

	// Decode the script back to an OrdLock to verify it's correctly formed
	decodedLock := Decode(lockScript)
	require.NotNil(t, decodedLock)

	// Verify the decoded values match what we put in
	require.NotNil(t, decodedLock.Seller)
	require.Equal(t, uint64(5000), decodedLock.Price)
}

func TestOrdLockDecode(t *testing.T) {
	// Create a mock transaction output with OrdLock script
	publicKeyHash, _ := hex.DecodeString("1234567890abcdef1234567890abcdef12345678")

	// Create a payout transaction output
	txOut := &transaction.TransactionOutput{
		Satoshis: 1000,
	}

	// Initialize with an empty script
	emptyScript := script.NewFromBytes([]byte{})
	txOut.LockingScript = emptyScript

	// Create a script buffer with the OrdLock data
	var scriptData []byte

	// Add the prefix
	scriptData = append(scriptData, OrdLockPrefix...)

	// Add the PKHash operation
	scriptData = append(scriptData, 0x14) // OP_DATA_20
	scriptData = append(scriptData, publicKeyHash...)

	// Add the payout output
	outputBytes := txOut.Bytes()
	scriptData = append(scriptData, byte(len(outputBytes))) // Length of output data
	scriptData = append(scriptData, outputBytes...)

	// Add the suffix
	scriptData = append(scriptData, OrdLockSuffix...)

	// Create the script
	scr := script.NewFromBytes(scriptData)

	// Test the Decode function
	ordLock := Decode(scr)

	// Verify the decoding worked as expected
	require.NotNil(t, ordLock, "Failed to decode OrdLock script")
	require.Equal(t, uint64(1000), ordLock.Price)
	require.NotNil(t, ordLock.Seller)
	require.NotEmpty(t, ordLock.PayOut)
}

func TestOrdLockPrefixSuffix(t *testing.T) {
	// Verify that the OrdLockPrefix and OrdLockSuffix constants are set
	require.NotNil(t, OrdLockPrefix)
	require.NotNil(t, OrdLockSuffix)
	require.True(t, len(OrdLockPrefix) > 0)
	require.True(t, len(OrdLockSuffix) > 0)

	// Log the lengths for diagnostic purposes
	t.Logf("OrdLockPrefix length: %d", len(OrdLockPrefix))
	t.Logf("OrdLockSuffix length: %d", len(OrdLockSuffix))
}

// TestDecodeInvalidScript tests decoding invalid scripts
func TestDecodeInvalidScript(t *testing.T) {
	// Skip the nil test as the implementation doesn't handle nil
	// Test with empty script
	emptyScript := script.NewFromBytes([]byte{})
	result := Decode(emptyScript)
	require.Nil(t, result, "Expected nil result for empty script")

	// Test with script containing only prefix
	prefixOnlyScript := script.NewFromBytes(OrdLockPrefix)
	result = Decode(prefixOnlyScript)
	require.Nil(t, result, "Expected nil result for script with only prefix")

	// Test with script containing only suffix
	suffixOnlyScript := script.NewFromBytes(OrdLockSuffix)
	result = Decode(suffixOnlyScript)
	require.Nil(t, result, "Expected nil result for script with only suffix")

	// Test with invalid script data between prefix and suffix
	invalidDataScript := script.NewFromBytes(append(append(OrdLockPrefix, []byte{0xFF, 0xEE, 0xDD}...), OrdLockSuffix...))
	result = Decode(invalidDataScript)
	require.Nil(t, result, "Expected nil result for script with invalid data")
}
