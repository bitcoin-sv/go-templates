package shrug

import (
	"math/big"

	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/script/interpreter"
)

const SHRUG_TAG = "¯\\_(ツ)_/¯"

type Shrug struct {
	Id           *overlay.Outpoint
	Amount       *big.Int
	ScriptSuffix []byte
}

func Decode(s *script.Script) *Shrug {
	shrug := &Shrug{}
	pos := 0
	if op, err := s.ReadOp(&pos); err != nil {
		return nil
	} else if len(op.Data) != 9 || string(op.Data[:8]) != SHRUG_TAG {
		return nil
	} else if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if len(op.Data) == 36 {
		shrug.Id = overlay.NewOutpointFromBytes([36]byte(op.Data))
	} else if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if op.Op != script.Op2DROP {
		return nil
	} else if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if number, err := interpreter.MakeScriptNumber(op.Data, len(op.Data), true, true); err != nil {
		return nil
	} else {
		shrug.Amount = number.Val
	}

	if op, err := s.ReadOp(&pos); err != nil {
		return nil
	} else if op.Op != script.OpDROP {
		return nil
	}
	shrug.ScriptSuffix = (*s)[pos:]
	return shrug
}

func (i *Shrug) Lock() *script.Script {
	s := &script.Script{}
	_ = s.AppendPushData([]byte(SHRUG_TAG))
	if i.Id != nil {
		_ = s.AppendPushData(i.Id.Bytes())
	} else {
		_ = s.AppendOpcodes(script.Op0)
	}
	_ = s.AppendOpcodes(script.Op2DROP)
	if i.Amount != nil {
		_ = s.AppendPushData((&interpreter.ScriptNumber{
			Val:          i.Amount,
			AfterGenesis: true,
		}).Bytes())
	} else {
		_ = s.AppendOpcodes(script.Op0)
	}
	_ = s.AppendOpcodes(script.OpDROP)
	return script.NewFromBytes(append(*s, i.ScriptSuffix...))
}
