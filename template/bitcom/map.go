package bitcom

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/bsv-blockchain/go-sdk/script"
)

const MapPrefix = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"

type MapCmd string

var ZERO = 0

var (
	MapCmdSet    MapCmd = "SET"
	MapCmdDel    MapCmd = "DEL"
	MapCmdAdd    MapCmd = "ADD"
	MapCmdSelect MapCmd = "SELECT"
)

type Map struct {
	Cmd  MapCmd            `json:"cmd"`
	Data map[string]string `json:"data"`
	Adds []string          `json:"adds,omitempty"`
}

// DecodeMap decodes the map data from the transaction script
func DecodeMap(scr script.Script) *Map {
	pos := &ZERO
	var op *script.ScriptChunk
	var err error

	if op, err = scr.ReadOp(pos); err != nil {
		return nil
	}

	m := &Map{
		Cmd:  MapCmd(op.Data),
		Data: make(map[string]string),
	}

	if m.Cmd == MapCmdSet {
		for {
			// Save position to revert if needed
			keyPos := *pos

			// Try to read key
			if op, err = scr.ReadOp(pos); err != nil {
				break
			}
			opKey := strings.Replace(string(bytes.Replace(op.Data, []byte{0}, []byte{' '}, -1)), "\\u0000", " ", -1)

			// Try to read value
			if op, err = scr.ReadOp(pos); err != nil {
				// Couldn't read value, revert to position before key and break
				*pos = keyPos
				break
			}

			if !utf8.Valid([]byte(opKey)) || !utf8.Valid(op.Data) {
				continue
			}

			m.Data[opKey] = strings.Replace(string(bytes.Replace(op.Data, []byte{0}, []byte{' '}, -1)), "\\u0000", " ", -1)
		}
	}
	return m
}
