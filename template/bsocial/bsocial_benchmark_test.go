package bsocial

import (
	"os"
	"testing"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

// loadTransactionForBenchmark loads a real transaction from the test data
func loadTransactionForBenchmark(b *testing.B, txID string) *transaction.Transaction {
	b.Helper()

	// Construct the file path from the txID
	filePath := "testdata/" + txID + ".hex"

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("Failed to read transaction file '%s': %v", filePath, err)
		return nil
	}

	// Parse raw transaction
	rawTx := string(data)
	tx, err := transaction.NewTransactionFromHex(rawTx)
	if err != nil {
		b.Fatalf("Failed to parse raw transaction: %v", err)
		return nil
	}

	return tx
}

// setupTestTransaction creates a transaction with varying numbers of outputs for benchmarking
func setupTestTransaction(b *testing.B, numOutputs int, includePrefix bool) *transaction.Transaction {
	b.Helper()
	tx := transaction.NewTransaction()

	// Add one B protocol output with data
	bScript := &script.Script{}
	if includePrefix {
		_ = bScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	} else {
		_ = bScript.AppendOpcodes(script.OpRETURN)
	}
	_ = bScript.AppendPushData([]byte(bitcom.BPrefix))
	_ = bScript.AppendPushData([]byte("This is test content"))
	_ = bScript.AppendPushData([]byte(string(bitcom.MediaTypeTextPlain)))
	_ = bScript.AppendPushData([]byte(string(bitcom.EncodingUTF8)))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: bScript,
		Satoshis:      0,
	})

	// Add one MAP protocol output that properly defines a post
	mapScript := &script.Script{}
	if includePrefix {
		_ = mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	} else {
		_ = mapScript.AppendOpcodes(script.OpRETURN)
	}
	_ = mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	_ = mapScript.AppendPushData([]byte("SET"))
	_ = mapScript.AppendPushData([]byte("app"))
	_ = mapScript.AppendPushData([]byte(AppName))
	_ = mapScript.AppendPushData([]byte("type"))
	_ = mapScript.AppendPushData([]byte(string(TypePostReply)))
	// Add content for MAP to ensure post is created
	_ = mapScript.AppendPushData([]byte("content"))
	_ = mapScript.AppendPushData([]byte("Benchmark post content"))
	// Add media type
	_ = mapScript.AppendPushData([]byte("mediaType"))
	_ = mapScript.AppendPushData([]byte(string(bitcom.MediaTypeTextPlain)))
	// Add encoding
	_ = mapScript.AppendPushData([]byte("encoding"))
	_ = mapScript.AppendPushData([]byte(string(bitcom.EncodingUTF8)))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	// Add additional dummy outputs if requested
	for i := 0; i < numOutputs-2; i++ {
		dummyScript := &script.Script{}
		_ = dummyScript.AppendOpcodes(script.OpDUP, script.OpHASH160)
		_ = dummyScript.AppendPushData([]byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78})
		_ = dummyScript.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIG)

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: dummyScript,
			Satoshis:      1000,
		})
	}

	return tx
}

// BenchmarkDecodeTransaction runs performance benchmarks on the DecodeTransaction function
func BenchmarkDecodeTransaction(b *testing.B) {
	// These are the real transaction IDs from the test vectors
	benchCases := []struct {
		name string
		txID string
	}{
		{"Post_Basic_1", "266c2a52d7d1f30709c847424d8195eeef8a0172f190be6244e5c8a1c2e44d94"},
		{"Post_Basic_2", "38c914d2c47c2ff063cf9f5705e3ceaa557aca4092ed5047177d5e8f913c0b69"},
		{"Reply_Basic", "8ca367aadc788d4f792b78f10577427840f2c31aae7cf9ffec9b327a79c883ef"},
		{"Like_Basic", "e89cd18de70bab82ccbea0836805b0039b61728f0641d89e8834d5225a593419"},
	}

	// Load each transaction and run the benchmark
	for _, bc := range benchCases {
		tx := loadTransactionForBenchmark(b, bc.txID)
		if tx == nil {
			b.Fatalf("Failed to load transaction for benchmark: %s", bc.name)
		}

		b.Run(bc.name, func(b *testing.B) {
			// Reset the timer for setup code
			b.ResetTimer()

			// Run the DecodeTransaction function b.N times
			for i := 0; i < b.N; i++ {
				_ = DecodeTransaction(tx)
			}
		})
	}

	// Also compare with synthetic transactions of different sizes
	syntheticCases := []struct {
		name          string
		numOutputs    int
		includePrefix bool
	}{
		{"Synthetic_Simple", 2, true},
		{"Synthetic_Medium", 5, true},
		{"Synthetic_Large", 10, true},
	}

	for _, sc := range syntheticCases {
		tx := setupTestTransaction(b, sc.numOutputs, sc.includePrefix)

		b.Run(sc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = DecodeTransaction(tx)
			}
		})
	}
}

// BenchmarkParseRawBData benchmarks the raw B data parsing
func BenchmarkParseRawBData(b *testing.B) {
	benchCases := []struct {
		name   string
		script []byte
	}{
		{"Small", []byte{0x00, 0x6a, 0x21, 0x31, 0x39, 0x48, 0x78, 0x69, 0x67, 0x56, 0x34, 0x51, 0x79, 0x42, 0x76, 0x33, 0x74, 0x48, 0x70, 0x51, 0x56, 0x63, 0x55, 0x45, 0x51, 0x79, 0x71, 0x31, 0x70, 0x7a, 0x5a, 0x56, 0x64, 0x6f, 0x41, 0x75, 0x74, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x0d, 0x74, 0x65, 0x78, 0x74, 0x2f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x05, 0x75, 0x74, 0x66, 0x2d, 0x38}},
		{"Medium", []byte{0x00, 0x6a, 0x21, 0x31, 0x39, 0x48, 0x78, 0x69, 0x67, 0x56, 0x34, 0x51, 0x79, 0x42, 0x76, 0x33, 0x74, 0x48, 0x70, 0x51, 0x56, 0x63, 0x55, 0x45, 0x51, 0x79, 0x71, 0x31, 0x70, 0x7a, 0x5a, 0x56, 0x64, 0x6f, 0x41, 0x75, 0x74, 0x20, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x20, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x0d, 0x74, 0x65, 0x78, 0x74, 0x2f, 0x70, 0x6c, 0x61, 0x69, 0x6e, 0x05, 0x75, 0x74, 0x66, 0x2d, 0x38}},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Convert to Script for benchmarking
			s := script.NewFromBytes(bc.script)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Replace with proper call to decode B data from script
				bitcom.DecodeB(s)
			}
		})
	}
}

// BenchmarkProcessMapData benchmarks the processMapData function
func BenchmarkProcessMapData(b *testing.B) {
	// Create test map data for different BSocial types
	postMap := &bitcom.Map{
		Data: map[string]string{
			"app":       AppName,
			"type":      string(TypePostReply),
			"content":   "Test post content",
			"mediaType": string(bitcom.MediaTypeTextPlain),
			"encoding":  string(bitcom.EncodingUTF8),
		},
	}

	likeMap := &bitcom.Map{
		Data: map[string]string{
			"app":  AppName,
			"type": string(TypeLike),
			"tx":   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	replyMap := &bitcom.Map{
		Data: map[string]string{
			"app":       AppName,
			"type":      string(TypePostReply),
			"context":   string(ContextTx),
			"tx":        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"content":   "Test reply content",
			"mediaType": string(bitcom.MediaTypeTextPlain),
			"encoding":  string(bitcom.EncodingUTF8),
		},
	}

	followMap := &bitcom.Map{
		Data: map[string]string{
			"app":   AppName,
			"type":  string(TypeFollow),
			"bapID": "test-user-id",
		},
	}

	benchCases := []struct {
		name    string
		mapData *bitcom.Map
	}{
		{"Post", postMap},
		{"Like", likeMap},
		{"Reply", replyMap},
		{"Follow", followMap},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				bsocial := &BSocial{}
				processMapData(bc.mapData, bsocial)
			}
		})
	}
}
