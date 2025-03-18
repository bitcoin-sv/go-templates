package bitcom

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/bsv-blockchain/go-sdk/script"
)

const MapPrefix = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"

type MapCmd string

var (
	MapCmdSet    MapCmd = "SET"
	MapCmdDel    MapCmd = "DEL"
	MapCmdAdd    MapCmd = "ADD"
	MapCmdSelect MapCmd = "SELECT"
)

type Map struct {
	Cmd  MapCmd            `json:"cmd"`
	Data map[string]string `json:"data"`
}

func DecodeMap(bitcom *Bitcom) []*Map {
	maps := []*Map{}
	for _, proto := range bitcom.Protocols {
		if proto.Protocol == MapPrefix {
			pos := &proto.Pos
			scr := script.NewFromBytes(proto.Script)
			if op, err := scr.ReadOp(pos); err != nil {
				return maps
			} else {
				m := &Map{
					Cmd:  MapCmd(op.Data),
					Data: make(map[string]string),
				}
				for {
					prevIdx := *pos
					op, err = scr.ReadOp(pos)
					opKey := strings.Replace(string(bytes.Replace(op.Data, []byte{0}, []byte{' '}, -1)), "\\u0000", " ", -1)
					prevIdx = *pos
					if op, err = scr.ReadOp(pos); err != nil {
						*pos = prevIdx
						break
					}

					if !utf8.Valid([]byte(opKey)) || !utf8.Valid(op.Data) {
						continue
					}

					m.Data[opKey] = strings.Replace(string(bytes.Replace(op.Data, []byte{0}, []byte{' '}, -1)), "\\u0000", " ", -1)
				}
				maps = append(maps, m)
			}
		}
	}
	return maps
}
