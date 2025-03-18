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
	for _, proto := range b.Protocols {
		if proto.Protocol == AIPPrefix {

			pos := &proto.Pos
			scr := script.NewFromBytes(proto.Script)

			aip := &AIP{}

			// Read ALGORITHM
			if op, err := scr.ReadOp(pos); err != nil {
				continue
			} else {
				aip.Algorithm = string(op.Data)
			}

			// Read ADDRESS
			if op, err := scr.ReadOp(pos); err != nil {
				continue
			} else {
				aip.Address = string(op.Data)
			}

			// Read SIGNATURE
			if op, err := scr.ReadOp(pos); err != nil {
				continue
			} else {
				aip.Signature = string(op.Data)
			}

			// Read optional FIELD INDEXES
			// If present, these indicate which fields were signed
			for {
				if op, err := scr.ReadOp(pos); err != nil {
					break
				} else {
					index, err := strconv.Atoi(string(op.Data))
					if err != nil {
						break
					}

					aip.FieldIndexes = append(aip.FieldIndexes, index)
				}
			}

			aips = append(aips, aip)
		}
	}
	return aips
}
