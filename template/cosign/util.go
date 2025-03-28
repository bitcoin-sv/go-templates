package cosign

import (
	"encoding/hex"

	"github.com/bsv-blockchain/go-sdk/script"
)

type Cosign struct {
	Address  string `json:"address"`
	Cosigner string `json:"cosigner"`
}

func Decode(s *script.Script) *Cosign {
	chunks, _ := s.Chunks()
	for i := range len(chunks) - 6 {
		if chunks[0+i].Op == script.OpDUP &&
			chunks[1+i].Op == script.OpHASH160 &&
			len(chunks[2+i].Data) == 20 &&
			chunks[3+i].Op == script.OpEQUALVERIFY &&
			chunks[4+i].Op == script.OpCHECKSIGVERIFY &&
			len(chunks[5+i].Data) == 33 &&
			chunks[6+i].Op == script.OpCHECKSIG {

			cosign := &Cosign{
				Cosigner: hex.EncodeToString(chunks[5+i].Data),
			}
			if add, err := script.NewAddressFromPublicKeyHash(chunks[2+i].Data, true); err == nil {
				cosign.Address = add.AddressString
			}
			return cosign
		}
	}
	return nil
}
