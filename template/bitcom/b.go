package bitcom

import (
	"github.com/bsv-blockchain/go-sdk/script"
)

// B PROTOCOL - PREFIX DATA MEDIA_TYPE ENCODING FILENAME

// BPrefix is the bitcom protocol prefix for B
const BPrefix = "19HxigV4QyBv3tHpQVcUEQyq1pzZVdoAut"

// Media types
type MediaType string

const (
	MediaTypeTextPlain    MediaType = "text/plain"
	MediaTypeTextMarkdown MediaType = "text/markdown"
	MediaTypeTextHTML     MediaType = "text/html"
	MediaTypeImagePNG     MediaType = "image/png"
	MediaTypeImageJPEG    MediaType = "image/jpeg"
)

type Encoding string

var (
	EncodingUTF8  Encoding = "utf-8"
	EncodingBinay Encoding = "binary"
)

// B represents B protocol data
type B struct {
	MediaType MediaType `json:"mediaType"`
	Encoding  Encoding  `json:"encoding"`
	Data      []byte    `json:"data"`
	Filename  string    `json:"filename,omitempty"`
}

// DecodeB decodes the b data from the transaction script
func DecodeB(b *Bitcom) []*B {
	bs := []*B{}
	for _, proto := range b.Protocols {
		if proto.Protocol == BPrefix {
			pos := &proto.Pos
			scr := script.NewFromBytes(proto.Script)
			var op *script.ScriptChunk
			var err error

			b := &B{}

			// Protocol order: PREFIX DATA MEDIA_TYPE ENCODING FILENAME
			// Skip prefix as it's already checked

			// Read DATA
			if op, err = scr.ReadOp(pos); err != nil {
				continue
			}
			b.Data = op.Data

			// Read MEDIA_TYPE
			if op, err = scr.ReadOp(pos); err != nil {
				continue
			}
			b.MediaType = MediaType(op.Data)

			// Read ENCODING
			if op, err = scr.ReadOp(pos); err != nil {
				continue
			}
			b.Encoding = Encoding(op.Data)

			// Try to read optional FILENAME
			if op, err = scr.ReadOp(pos); err == nil {
				// Successfully read filename
				b.Filename = string(op.Data)
			}

			// Add to results
			bs = append(bs, b)

		}
	}
	return bs
}
