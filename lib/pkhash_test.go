package lib

import (
	"encoding/json"
	"testing"
)

func TestPKHashAddress(t *testing.T) {
	p := PKHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	addr := p.Address(Mainnet)
	if addr == "" {
		t.Error("expected non-empty address")
	}
}

func TestPKHashMarshalUnmarshalJSON(t *testing.T) {
	p := PKHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var p2 PKHash
	err = json.Unmarshal(b, &p2)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(p2) != 20 {
		t.Errorf("expected length 20, got %d", len(p2))
	}
}

func TestPKHashFromAddress(t *testing.T) {
	p := PKHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	addr := p.Address(Mainnet)
	var p2 PKHash
	err := p2.FromAddress(addr)
	if err != nil {
		t.Fatalf("FromAddress error: %v", err)
	}
	if len(p2) != 20 {
		t.Errorf("expected length 20, got %d", len(p2))
	}
	// Invalid address
	err = p2.FromAddress("notanaddress")
	if err == nil {
		t.Error("expected error for invalid address")
	}
}
