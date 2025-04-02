package bitcom

import (
	"strconv"

	"github.com/bsv-blockchain/go-sdk/script"
)

// AIPPrefix is the bitcom protocol prefix for AIP
const AIPPrefix = "15PciHG22SNLQJXMoSUaWVi7WSqc7hCfva"

// AIP represents an AIP
type AIP struct {
	Algorithm    string `json:"algorithm"`
	Address      string `json:"address"`
	Signature    string `json:"signature"`
	FieldIndexes []int  `json:"fieldIndexes,omitempty"`
}

// DecodeAIP decodes the AIP data from the transaction script
func DecodeAIP(b *Bitcom) []*AIP {
	aips := []*AIP{}

	// Safety check for nil
	if b == nil || len(b.Protocols) == 0 {
		return aips
	}

	for _, proto := range b.Protocols {
		if proto.Protocol == AIPPrefix {
			scr := script.NewFromBytes(proto.Script)
			if scr == nil {
				continue
			}

			// Parse script into chunks
			chunks, err := scr.Chunks()
			if err != nil || len(chunks) < 3 { // Need at least algorithm, address, and signature
				continue
			}

			aip := &AIP{}

			// Read ALGORITHM (first chunk)
			if len(chunks) > 0 {
				aip.Algorithm = string(chunks[0].Data)
			} else {
				continue
			}

			// Read ADDRESS (second chunk)
			if len(chunks) > 1 {
				aip.Address = string(chunks[1].Data)
			} else {
				continue
			}

			// Read SIGNATURE (third chunk)
			if len(chunks) > 2 {
				aip.Signature = string(chunks[2].Data)
			} else {
				continue
			}

			// Read optional FIELD INDEXES (remaining chunks)
			// If present, these indicate which fields were signed
			for i := 3; i < len(chunks); i++ {
				index, err := strconv.Atoi(string(chunks[i].Data))
				if err != nil {
					break // Stop if we encounter non-numeric data
				}
				aip.FieldIndexes = append(aip.FieldIndexes, index)
			}

			aips = append(aips, aip)
		}
	}

	return aips
}
