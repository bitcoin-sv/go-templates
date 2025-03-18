package bitcom

import "github.com/bsv-blockchain/go-sdk/script"

type Bitcom struct {
	Protocol string
	Script   []byte
	Pos      int
}

func Decode(scr *script.Script) (bitcoms []*Bitcom) {
	pos := findReturn(scr, 0)
	if pos == -1 {
		return
	}
	pos++

	for {
		pipePos := findPipe(scr, pos)
		bitcom := &Bitcom{
			Pos: pos,
		}
		if op, err := scr.ReadOp(&pos); err != nil {
			return
		} else {
			bitcom.Protocol = string(op.Data)
		}
		if pipePos == -1 {
			bitcom.Script = (*scr)[pos:]
			bitcoms = append(bitcoms, bitcom)
			return bitcoms
		}
		bitcom.Script = (*scr)[pos:]
		bitcoms = append(bitcoms, bitcom)
		pos = pipePos + 2
	}
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
