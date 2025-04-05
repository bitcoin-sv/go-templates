package opns

import (
	"bytes"

	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/script"
)

var GENESIS, _ = overlay.NewOutpointFromString("58b7558ea379f24266c7e2f5fe321992ad9a724fd7a87423ba412677179ccb25_0")
var DIFFICULTY = 22

type OpNS struct {
	Claimed []byte `json:"claimed,omitempty"`
	Domain  string `json:"domain"`
	PoW     []byte `json:"pow,omitempty"`
}

func Decode(scr *script.Script) *OpNS {
	if opNSPrefixIndex := bytes.Index(*scr, OpNSPrefix); opNSPrefixIndex == -1 {
		return nil
	} else if opNSSuffixIndex := bytes.Index(*scr, OpNSSuffix); opNSSuffixIndex == -1 {
		return nil
	} else {
		opNS := &OpNS{}
		stateScript := script.NewFromBytes((*scr)[opNSSuffixIndex+len(OpNSSuffix)+2:])
		pos := 0
		if genesisOp, err := stateScript.ReadOp(&pos); err != nil {
			return nil
		} else if genesisOp.Op != 36 {
			return nil
		} else if !overlay.NewOutpointFromTxBytes([36]byte(genesisOp.Data)).Equal(GENESIS) {
			return nil
		} else if claimedOp, err := stateScript.ReadOp(&pos); err != nil {
			return nil
		} else if domainOp, err := stateScript.ReadOp(&pos); err != nil {
			return nil
		} else if powOp, err := stateScript.ReadOp(&pos); err != nil {
			return nil
		} else {
			opNS.Claimed = claimedOp.Data
			opNS.Domain = string(domainOp.Data)
			opNS.PoW = powOp.Data
		}
		return opNS
	}
}
