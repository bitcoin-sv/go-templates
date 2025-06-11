package bsocial

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/stretchr/testify/require"
)

// Error message templates for consistent error reporting and to avoid linter warnings
const (
	// General test error formats
	ErrMsgActionNotFound   = "Expected to find a %s action for test vector '%s', but none was found"
	ErrMsgActionFound      = "Expected not to find a %s action for test vector '%s', but found one"
	ErrMsgNilForTestVector = "Expected %s to be nil for test vector '%s'"
	ErrMsgDecodeFailure    = "Failed to decode %s for test vector '%s'"

	// Value comparison error formats
	ErrMsgWrongType        = "Expected action type '%s' but found '%s' for test vector '%s'"
	ErrMsgWrongContextType = "Expected context_type '%s' but got '%s' for test vector '%s'"
	ErrMsgWrongPostTx      = "Expected post_tx '%s' but got '%s' for test vector '%s'"
	ErrMsgWrongBapID       = "Expected bap_id '%s' but got '%s' for test vector '%s'"
	ErrMsgWrongChannel     = "Expected channel '%s' but got '%s' for test vector '%s'"
	ErrMsgWrongContent     = "Expected content '%s' but got '%s' for test vector '%s'"
)

// TestVector represents a single BSocial test case with expected outcomes
type TestVector struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	RawTransaction string         `json:"raw_transaction,omitempty"`
	Expected       map[string]any `json:"expected"`
}

// TestVectors represents a collection of BSocial test cases
type TestVectors struct {
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Vectors     []TestVector `json:"vectors"`
}

// LoadTestVectors loads and parses test vectors from a JSON file
func LoadTestVectors(t *testing.T, filePath string) TestVectors {
	t.Helper()

	// Read test vectors file
	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read test vectors file: %s", filePath)

	// Parse test vectors
	var vectors TestVectors
	err = json.Unmarshal(data, &vectors)
	require.NoError(t, err, "Failed to parse test vectors")

	return vectors
}

// GetTransactionFromVector loads a transaction from a file based on the txid in the test vector
func GetTransactionFromVector(t *testing.T, vector TestVector) *transaction.Transaction {
	t.Helper()

	// Get transaction ID from expected values
	var txID string
	if id, ok := vector.Expected["tx_id"].(string); ok && id != "" {
		txID = id
	} else {
		t.Logf("No tx_id found in expected values for test vector '%s'", vector.Name)
		return nil
	}

	// Construct the file path from the txID
	filePath := "testdata/" + txID + ".hex"
	t.Logf("Attempting to read transaction from file: %s", filePath)

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Logf("Failed to read transaction file '%s': %v", filePath, err)
		return nil
	}

	// Clean up the hex data
	rawTx := string(data)
	rawTx = strings.TrimSuffix(rawTx, "%")  // Remove trailing %
	rawTx = strings.TrimSuffix(rawTx, "\n") // Remove newline
	t.Logf("Read transaction hex from file, length: %d characters", len(rawTx))

	// Skip if empty
	if rawTx == "" {
		t.Skipf("Skipping test vector '%s' because raw transaction is empty", vector.Name)
		return nil
	}

	// Log the beginning of the transaction hex for diagnostic purposes
	if len(rawTx) > 50 {
		t.Logf("Transaction hex starts with: %s...", rawTx[:50])
	} else {
		t.Logf("Transaction hex: %s", rawTx)
	}

	// Parse raw transaction
	tx, err := transaction.NewTransactionFromHex(rawTx)
	if err != nil {
		t.Errorf("Failed to parse raw transaction for test vector '%s': %v", vector.Name, err)
		return nil
	}

	// Log transaction structure for diagnostic purposes
	t.Logf("Transaction parsed successfully. Number of outputs: %d", len(tx.Outputs))
	for i, output := range tx.Outputs {
		if output.LockingScript != nil {
			t.Logf("Output #%d: Has LockingScript with length %d", i, len(output.LockingScript.String()))
		} else {
			t.Logf("Output #%d: LockingScript is nil", i)
		}
	}

	return tx
}
