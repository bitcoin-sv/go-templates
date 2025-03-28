package bsv21

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/bitcoin-sv/go-templates/lib"
	"github.com/bitcoin-sv/go-templates/template/inscription"
	"github.com/bsv-blockchain/go-sdk/script"
)

type Op string

var (
	OpMint     Op = "deploy+mint"
	OpTransfer Op = "transfer"
	OpBurn     Op = "burn"
)

type Bsv21 struct {
	Id       string  `json:"id,omitempty"`
	Op       string  `json:"op"`
	Symbol   *string `json:"sym,omitempty"`
	Decimals uint8   `json:"dec"`
	Icon     *string `json:"icon,omitempty"`
	Amt      uint64  `json:"amt"`
	Insc     *inscription.Inscription
}

func Decode(scr *script.Script) *Bsv21 {
	insc := inscription.Decode(scr)
	data := map[string]string{}
	if insc == nil {
		return nil
	} else if insc.File.Type != "application/bsv-20" {
		return nil
	} else if err := json.Unmarshal(insc.File.Content, &data); err != nil {
		return nil
	} else if p, ok := data["p"]; !ok || p != "bsv-20" {
		return nil
	} else {
		bsv21 := &Bsv21{
			Insc: insc,
		}
		if op, ok := data["op"]; ok {
			bsv21.Op = strings.ToLower(op)
		} else {
			return nil
		}

		if amt, ok := data["amt"]; ok {
			if bsv21.Amt, err = strconv.ParseUint(amt, 10, 64); err != nil {
				return nil
			}
		}

		if dec, ok := data["dec"]; ok {
			var val uint64
			if val, err = strconv.ParseUint(dec, 10, 8); err != nil || val > 18 {
				return nil
			}
			bsv21.Decimals = uint8(val)
		}

		switch bsv21.Op {
		case string(OpMint):
			if sym, ok := data["sym"]; ok {
				bsv21.Symbol = &sym
			}
			if icon, ok := data["icon"]; ok {
				bsv21.Icon = &icon
			}
		case string(OpTransfer), string(OpBurn):
			if id, ok := data["id"]; !ok {
				return nil
			} else if _, err = lib.NewOutpointFromString(id); err != nil {
				return nil
			} else {
				bsv21.Id = id
			}
		default:
			return nil
		}
		return bsv21
	}
}
