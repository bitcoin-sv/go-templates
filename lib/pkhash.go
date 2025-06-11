package lib

import (
	"encoding/json"

	"github.com/bsv-blockchain/go-sdk/script"
)

// Network represents a BSV network
type Network int

const (
	Mainnet Network = 0
	Testnet Network = 1
)

// PKHash is a wrapper around a byte slice representing a public key hash
type PKHash []byte

// Address returns the address string representation of the public key hash
func (p *PKHash) Address(network ...Network) string {
	mainnet := true
	if len(network) > 0 {
		mainnet = network[0] != Testnet
	}
	add, _ := script.NewAddressFromPublicKeyHash(*p, mainnet)
	return add.AddressString
}

// MarshalJSON serializes PKHash to its address representation
func (p PKHash) MarshalJSON() ([]byte, error) {
	add := p.Address()
	return json.Marshal(add)
}

// FromAddress creates a PKHash from an address string
func (p *PKHash) FromAddress(a string) error {
	if add, err := script.NewAddressFromString(a); err != nil {
		return err
	} else {
		*p = PKHash(add.PublicKeyHash)
	}
	return nil
}

// UnmarshalJSON deserializes an address string to PKHash
func (p *PKHash) UnmarshalJSON(data []byte) error {
	var add string
	err := json.Unmarshal(data, &add)
	if err != nil {
		return err
	}
	return p.FromAddress(add)
}
