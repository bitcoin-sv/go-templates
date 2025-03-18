package bitcom

import "github.com/bsv-blockchain/go-sdk/script"

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
		s.AppendOpcodes(script.OpRETURN)
		for i, p := range b.Protocols {
			s.AppendPushData([]byte(p.Protocol))
			s.AppendPushData(p.Script)
			if i < len(b.Protocols)-1 {
				s.AppendPushData([]byte("|"))
			}
		}
	}
	return s
}

func findReturn(scr *script.Script, pos int) int {
	for i := pos; i < len(*scr); i++ {
		if op, err := scr.ReadOp(&i); err == nil && op.Op == script.OpRETURN {
			return i
		}
	}
	return -1
}

func findPipe(scr *script.Script, pos int) int {
	for i := pos; i < len(*scr); i++ {
		if op, err := scr.ReadOp(&i); err == nil && op.Op == script.OpDATA1 && op.Data[0] == '|' {
			return i
		}
	}
	return -1
}
