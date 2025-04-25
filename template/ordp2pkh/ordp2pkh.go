// Package ordp2pkh provides functionality for creating and decoding Bitcoin scripts
// that combine Ordinal inscriptions with standard P2PKH locking scripts.
//
// OrdP2PKH allows you to create scripts that both contain inscription data (like images,
// text, or other content) and are spendable using a standard P2PKH address. This enables
// the creation of Ordinal NFTs that can be transferred using standard Bitcoin transactions.
package ordp2pkh

import (
	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bitcoin-sv/go-templates/template/inscription"
	"github.com/bitcoin-sv/go-templates/template/p2pkh"
	"github.com/bsv-blockchain/go-sdk/script"
)

// OrdP2PKH represents an inscription with a P2PKH locking script
type OrdP2PKH struct {
	Inscription *inscription.Inscription `json:"inscription"`
	Address     *script.Address          `json:"address,omitempty"`
	Metadata    *bitcom.Map              `json:"metadata,omitempty"`
}

// Decode attempts to extract an OrdP2PKH from a script
func Decode(s *script.Script) *OrdP2PKH {
	if s == nil {
		return nil
	}

	// Try to decode the inscription first
	inscr := inscription.Decode(s)
	if inscr == nil {
		return nil
	}

	// This is a valid inscription, so now we need to find the P2PKH address
	// from either the prefix or suffix
	addr := getAddressFromScript(inscr)
	if addr == nil {
		// No valid P2PKH address found
		return nil
	}

	// Find MAP metadata in the script
	metadata := getMetadataFromScript(s)

	// Create the OrdP2PKH
	return &OrdP2PKH{
		Inscription: inscr,
		Address:     addr,
		Metadata:    metadata,
	}
}

// getMetadataFromScript attempts to extract MAP metadata from a script
func getMetadataFromScript(s *script.Script) *bitcom.Map {
	if s == nil {
		return nil
	}

	// Use bitcom.Decode to find all BitCom protocols in the script
	bc := bitcom.Decode(s)
	if bc == nil || len(bc.Protocols) == 0 {
		return nil
	}

	// Look for MAP protocol
	for _, proto := range bc.Protocols {
		if proto.Protocol == bitcom.MapPrefix {
			return bitcom.DecodeMap(proto.Script)
		}
	}

	return nil
}

// getAddressFromScript extracts a P2PKH address from an inscription's prefix or suffix
func getAddressFromScript(inscription *inscription.Inscription) *script.Address {
	// Check prefix first
	if len(inscription.ScriptPrefix) > 0 {
		prefix := script.NewFromBytes(inscription.ScriptPrefix)
		if address := p2pkh.Decode(prefix, true); address != nil {
			return address
		}
	}

	// Then check suffix
	if len(inscription.ScriptSuffix) > 0 {
		suffix := script.NewFromBytes(inscription.ScriptSuffix)
		if address := p2pkh.Decode(suffix, true); address != nil {
			return address
		}

		// If direct decode failed, check if a P2PKH script is at the beginning of a larger suffix script
		if addr := extractP2PKHFromScript(suffix); addr != nil {
			return addr
		}
	}

	// Finally check prefix with extraction method as well
	if len(inscription.ScriptPrefix) > 0 {
		prefix := script.NewFromBytes(inscription.ScriptPrefix)
		if addr := extractP2PKHFromScript(prefix); addr != nil {
			return addr
		}
	}

	return nil
}

// extractP2PKHFromScript attempts to extract a P2PKH address from a script
// that might have additional data after the P2PKH part
func extractP2PKHFromScript(s *script.Script) *script.Address {
	chunks, err := s.Chunks()
	if err != nil || len(chunks) < 5 {
		return nil
	}

	// Check for P2PKH pattern: OP_DUP OP_HASH160 <pubkeyhash> OP_EQUALVERIFY OP_CHECKSIG
	if chunks[0].Op == script.OpDUP &&
		chunks[1].Op == script.OpHASH160 &&
		len(chunks[2].Data) == 20 &&
		chunks[3].Op == script.OpEQUALVERIFY &&
		chunks[4].Op == script.OpCHECKSIG {

		// Create a standard P2PKH script with just the core components
		p2pkhScript := script.NewFromBytes([]byte{
			script.OpDUP,
			script.OpHASH160,
			script.OpDATA20,
		})

		// Append the pubkey hash (20 bytes)
		*p2pkhScript = append(*p2pkhScript, chunks[2].Data...)

		// Append the final opcodes
		*p2pkhScript = append(*p2pkhScript, script.OpEQUALVERIFY, script.OpCHECKSIG)

		// Use the standard p2pkh.Decode with the cleaned script
		return p2pkh.Decode(p2pkhScript, true)
	}

	return nil
}

// Lock creates a combined script that includes an inscription followed by a P2PKH locking script.
// Returns the combined script and any error encountered.
func (op *OrdP2PKH) Lock() (*script.Script, error) {
	return op.LockWithMapMetadata(nil)
}

// LockWithMapMetadata creates a combined script that includes an inscription, a P2PKH locking script,
// and optional MAP metadata.
// Returns the combined script and any error encountered.
func (op *OrdP2PKH) LockWithMapMetadata(metadata *bitcom.Map) (*script.Script, error) {
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

	// Generate the combined script with inscription and P2PKH
	combinedScript, err := op.Inscription.Lock()
	if err != nil {
		return nil, err
	}

	// If no metadata is provided, return the script as is
	if metadata == nil {
		return combinedScript, nil
	}

	// Validate MAP metadata - must have app and type fields
	if _, hasApp := metadata.Data["app"]; !hasApp {
		return combinedScript, nil // Return without MAP if app is missing
	}

	if _, hasType := metadata.Data["type"]; !hasType {
		return combinedScript, nil // Return without MAP if type is missing
	}

	// Create a standalone MAP script
	mapScript := &script.Script{}
	_ = mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = mapScript.AppendPushDataString(bitcom.MapPrefix)
	_ = mapScript.AppendPushDataString(string(metadata.Cmd))

	// Add all key-value pairs
	for key, value := range metadata.Data {
		_ = mapScript.AppendPushDataString(key)
		_ = mapScript.AppendPushDataString(value)
	}

	// Return the combined script with MAP script appended
	return script.NewFromBytes(append(*combinedScript, *mapScript...)), nil
}

// LockWithAddress is a convenience method that creates a new OrdP2PKH instance with the given address
// and inscription, then creates a combined script.
func LockWithAddress(address *script.Address, inscription *inscription.Inscription, metadata *bitcom.Map) (*script.Script, error) {
	op := &OrdP2PKH{
		Inscription: inscription,
		Address:     address,
	}
	return op.LockWithMapMetadata(metadata)
}
