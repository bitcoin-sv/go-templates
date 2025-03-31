// Package ordp2pkh provides functionality for creating and decoding Bitcoin scripts
// that combine Ordinal inscriptions with standard P2PKH locking scripts.
//
// OrdP2PKH allows you to create scripts that both contain inscription data (like images,
// text, or other content) and are spendable using a standard P2PKH address. This enables
// the creation of Ordinal NFTs that can be transferred using standard Bitcoin transactions.
package ordp2pkh

import (
	"github.com/bitcoin-sv/go-templates/template/inscription"
	"github.com/bitcoin-sv/go-templates/template/p2pkh"
	"github.com/bsv-blockchain/go-sdk/script"
)

// OrdP2PKH represents a combined Ordinal inscription and P2PKH script.
// It contains both the inscription data and the address for spending.
type OrdP2PKH struct {
	Inscription *inscription.Inscription
	Address     *script.Address
}

// Decode attempts to extract both an Inscription and a P2PKH address from a script.
// Returns nil if either component is not found in the script.
func Decode(scr *script.Script) *OrdP2PKH {
	// Try to decode the inscription first
	inscription := inscription.Decode(scr)
	if inscription == nil {
		return nil
	}

	// Check prefix first
	prefix := script.NewFromBytes(inscription.ScriptPrefix)
	if prefix != nil {
		address := p2pkh.Decode(prefix, true)
		if address != nil {
			return &OrdP2PKH{
				Inscription: inscription,
				Address:     address,
			}
		}
	}

	// Only check suffix if address wasn't found in the prefix
	suffix := script.NewFromBytes(inscription.ScriptSuffix)
	if suffix != nil {
		address := p2pkh.Decode(suffix, true)
		if address != nil {
			return &OrdP2PKH{
				Inscription: inscription,
				Address:     address,
			}
		}
	}

	// No valid address found
	return nil
}

// Lock creates a combined script that includes an inscription followed by a P2PKH locking script.
// Returns the combined script and any error encountered.
func (op *OrdP2PKH) Lock() (*script.Script, error) {
	// Create the P2PKH script
	p2pkhScript, err := p2pkh.Lock(op.Address)
	if err != nil {
		return nil, err
	}

	// Ensure we have a proper inscription
	if op.Inscription == nil {
		op.Inscription = &inscription.Inscription{}
	}

	// Set the P2PKH script as the suffix
	op.Inscription.ScriptSuffix = *p2pkhScript

	// Return the combined script
	return op.Inscription.Lock()
}
