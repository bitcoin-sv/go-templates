// Package sigma provides functionality for creating and verifying Sigma signatures
// for Bitcoin SV transactions. Sigma is a digital signature scheme for signing
// Bitcoin transaction data.
package sigma

import (
	"encoding/base64"
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bsv-blockchain/go-sdk/script"
)

// SIGMAPrefix is another recognized prefix in some implementations
const SIGMAPrefix = "SIGMA"

// SignatureAlgorithm represents the algorithm used for the signature
type SignatureAlgorithm string

const (
	// AlgoECDSA represents the ECDSA signature algorithm
	AlgoECDSA SignatureAlgorithm = "ECDSA"

	// AlgoSHA256ECDSA represents SHA256+ECDSA signature algorithm
	AlgoSHA256ECDSA SignatureAlgorithm = "SHA256-ECDSA"

	// AlgoBSM represents Bitcoin Signed Message algorithm
	AlgoBSM SignatureAlgorithm = "BSM"
)

// Sigma represents a Sigma signature
type Sigma struct {
	Algorithm      SignatureAlgorithm `json:"algorithm"`
	SignerAddress  string             `json:"signerAddress"`
	SignatureValue string             `json:"signatureValue"`
	Message        string             `json:"message,omitempty"`
	Nonce          string             `json:"nonce,omitempty"`
	VIN            int                `json:"vin,omitempty"`
}

// Decode decodes the Sigma data from the bitcom protocols
func Decode(b *bitcom.Bitcom) []*Sigma {
	signatures := []*Sigma{}

	// Safety check for nil
	if b == nil || len(b.Protocols) == 0 {
		return signatures
	}

	for i, proto := range b.Protocols {
		// Debug
		fmt.Printf("Protocol %d: %q, Script length: %d\n", i, proto.Protocol, len(proto.Script))
		fmt.Printf("SIGMAPrefix: %q\n", SIGMAPrefix)
		fmt.Printf("proto.Protocol == SIGMAPrefix: %v\n", proto.Protocol == SIGMAPrefix)
		fmt.Printf("Script hex: %x\n", proto.Script)

		// Check for SIGMA prefix
		if proto.Protocol == SIGMAPrefix {
			pos := 0 // Start from beginning of script
			scr := script.NewFromBytes(proto.Script)

			sigma := &Sigma{}

			// Debug
			fmt.Printf("Reading data from script...\n")

			// Read ALGORITHM - handle the case where it's prefixed with length
			if op, err := scr.ReadOp(&pos); err != nil {
				fmt.Printf("Error reading algorithm: %v\n", err)
				continue
			} else {
				// The algorithm field is prefixed with its length (03) for "BSM"
				if len(op.Data) > 1 && op.Data[0] == 0x03 {
					sigma.Algorithm = SignatureAlgorithm(string(op.Data[1:])) // Skip the length byte
				} else {
					sigma.Algorithm = SignatureAlgorithm(string(op.Data))
				}
				fmt.Printf("Algorithm: %q\n", sigma.Algorithm)
			}

			// Read SIGNER ADDRESS - handle the case where it's prefixed with quotes
			if op, err := scr.ReadOp(&pos); err != nil {
				fmt.Printf("Error reading signer address: %v\n", err)
				continue
			} else {
				if len(op.Data) > 1 && op.Data[0] == '"' {
					// If it starts with a quote, trim the quotes
					sigma.SignerAddress = string(op.Data[1 : len(op.Data)-1])
				} else {
					sigma.SignerAddress = string(op.Data)
				}
				fmt.Printf("SignerAddress: %q\n", sigma.SignerAddress)
			}

			// Read SIGNATURE VALUE
			if op, err := scr.ReadOp(&pos); err != nil {
				fmt.Printf("Error reading signature value: %v\n", err)
				continue
			} else {
				// Base64 encode the signature value
				sigma.SignatureValue = base64.StdEncoding.EncodeToString(op.Data)
				fmt.Printf("SignatureValue: %s\n", sigma.SignatureValue)
			}

			// Try to read optional fields
			if op, err := scr.ReadOp(&pos); err == nil {
				// This could be VIN or a type field (which we ignore)
				if string(op.Data) == "0" {
					sigma.VIN = 0
					fmt.Printf("VIN: %d\n", sigma.VIN)
				} else {
					// Skip the type field (which we don't need)
					// Read optional MESSAGE (what was signed)
					if op, err := scr.ReadOp(&pos); err == nil {
						sigma.Message = string(op.Data)
						fmt.Printf("Message: %q\n", sigma.Message)
					}

					// Read optional NONCE
					if op, err := scr.ReadOp(&pos); err == nil {
						sigma.Nonce = string(op.Data)
						fmt.Printf("Nonce: %q\n", sigma.Nonce)
					}
				}
			}

			fmt.Printf("Adding sigma to result\n")
			signatures = append(signatures, sigma)
		}
	}
	fmt.Printf("Returning %d signatures\n", len(signatures))
	return signatures
}

// GetSignatureBytes returns the signature as a byte array
func (s *Sigma) GetSignatureBytes() ([]byte, error) {
	if s.SignatureValue == "" {
		return nil, nil
	}

	// Signatures are always base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(s.SignatureValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 signature: %w", err)
	}

	return decoded, nil
}
