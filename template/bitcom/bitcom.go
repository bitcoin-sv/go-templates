package bitcom

import (
	"github.com/bsv-blockchain/go-sdk/script"
)

type Bitcom struct {
	Protocols    []*BitcomProtocol `json:"protos"`
	ScriptPrefix []byte            `json:"prefix,omitempty"`
}
type BitcomProtocol struct {
	Protocol string `json:"proto"`
	Script   []byte `json:"script"`
	Pos      int    `json:"pos"`
}

func Decode(scr *script.Script) (bitcom *Bitcom) {
	// Handle nil script safely
	if scr == nil {
		return &Bitcom{
			Protocols: []*BitcomProtocol{},
		}
	}

	pos := findReturn(scr, 0)
	if pos == -1 {
		return
	}
	bitcom = &Bitcom{
		ScriptPrefix: (*scr)[:pos-1],
	}
	pos++

	for {
		pipePos := findPipe(scr, pos)
		p := &BitcomProtocol{
			Pos: pos,
		}
		if op, err := scr.ReadOp(&pos); err != nil {
			return
		} else {
			p.Protocol = string(op.Data)
		}
		if pipePos == -1 {
			p.Script = (*scr)[pos:]
			bitcom.Protocols = append(bitcom.Protocols, p)
			return bitcom
		}
		p.Script = (*scr)[pos:]
		bitcom.Protocols = append(bitcom.Protocols, p)
		pos = pipePos + 2
	}
}

func (b *Bitcom) Lock() *script.Script {
	s := script.NewFromBytes(b.ScriptPrefix)
	if len(b.Protocols) > 0 {
		_ = s.AppendOpcodes(script.OpRETURN)
		for i, p := range b.Protocols {
			_ = s.AppendPushData([]byte(p.Protocol))
			_ = s.AppendPushData(p.Script)
			if i < len(b.Protocols)-1 {
				_ = s.AppendPushData([]byte("|"))
			}
		}
	}
	return s
}

func findReturn(scr *script.Script, pos int) int {
	// Handle nil script
	if scr == nil {
		return -1
	}

	for i := pos; i < len(*scr); i++ {
		if op, err := scr.ReadOp(&i); err == nil && op.Op == script.OpRETURN {
			return i
		}
	}
	return -1
}

func findPipe(scr *script.Script, pos int) int {
	// Handle nil script
	if scr == nil {
		return -1
	}

	for i := pos; i < len(*scr); i++ {
		if op, err := scr.ReadOp(&i); err == nil && op.Op == script.OpDATA1 && op.Data[0] == '|' {
			return i
		}
	}
	return -1
}

// ToScript converts a []byte to a script.Script or returns a script directly
// This is a helper function that can be used by all decoders
func ToScript(data any) *script.Script {
	switch d := data.(type) {
	case *script.Script:
		return d
	case script.Script:
		return &d
	case []byte:
		if d == nil {
			return nil
		}
		// Convert bytes to script
		s := script.NewFromBytes(d)
		return s
	default:
		return nil
	}
}
