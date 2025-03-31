package pow20

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/script/interpreter"
	"github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

type Pow20 struct {
	Txid          []byte         `json:"txid"`
	Vout          uint32         `json:"vout"`
	Symbol        string         `json:"sym"`
	Max           uint64         `json:"max"`
	Dec           uint8          `json:"dec"`
	Reward        uint64         `json:"cur"`
	Difficulty    uint8          `json:"dif"`
	Id            string         `json:"id"`
	Supply        uint64         `json:"sup"`
	LockingScript *script.Script `json:"lockingScript"`
}

type Pow20Unlocker struct {
	Pow20
	Nonce     []byte          `json:"nonce"`
	Recipient *script.Address `json:"recipient"`
}

func Decode(s *script.Script) *Pow20 {
	prefix := bytes.Index(*s, *pow20Prefix)
	if prefix == -1 {
		return nil
	}
	suffix := bytes.Index(*s, *pow20Suffix)
	if suffix == -1 {
		return nil
	}
	pos := prefix + len(*pow20Prefix)
	var err error
	var op *script.ScriptChunk

	p := &Pow20{}
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	}
	p.Symbol = string(op.Data)
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	}
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if number, err := interpreter.MakeScriptNumber(op.Data, len(op.Data), true, true); err != nil {
		return nil
	} else {
		p.Max = number.Val.Uint64()
	}
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if op.Op >= script.Op1 && op.Op <= script.Op16 {
		p.Dec = op.Op - 0x50
	} else if len(op.Data) == 1 {
		p.Dec = op.Data[0]
	}
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if number, err := interpreter.MakeScriptNumber(op.Data, len(op.Data), true, true); err != nil {
		return nil
	} else {
		p.Reward = number.Val.Uint64()
	}
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	}
	p.Difficulty = op.Op - 0x50

	pos = suffix + len(*pow20Suffix) + 2
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	}
	p.Id = string(op.Data)
	if op, err = s.ReadOp(&pos); err != nil {
		return nil
	} else if number, err := interpreter.MakeScriptNumber(op.Data, len(op.Data), true, true); err != nil {
		return nil
	} else {
		p.Supply = number.Val.Uint64()
	}
	p.LockingScript = s

	return p
}

func (p *Pow20) BuildUnlockTx(nonce []byte, recipient *script.Address, changeAddress *script.Address) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	unlock, err := p.Unlock(nonce, recipient)
	if err != nil {
		return nil, err
	}

	txid, _ := chainhash.NewHash(p.Txid)
	_ = tx.AddInputsFromUTXOs(&transaction.UTXO{
		TxID:                    txid,
		Vout:                    p.Vout,
		LockingScript:           p.LockingScript,
		Satoshis:                1,
		UnlockingScriptTemplate: unlock,
	})
	tx.Inputs[0].SequenceNumber = 0

	if p.Supply > p.Reward {
		restateScript := p.Lock(p.Supply - p.Reward)
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: restateScript,
			Satoshis:      1,
		})
	}
	rewardScript := BuildInscription(p.Id, p.Reward)
	_ = rewardScript.AppendOpcodes(script.OpDUP, script.OpHASH160)
	_ = rewardScript.AppendPushData(recipient.PublicKeyHash)
	_ = rewardScript.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIG)
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: rewardScript,
		Satoshis:      1,
	})
	if changeAddress != nil {
		change := &transaction.TransactionOutput{
			Change: true,
		}
		change.LockingScript, _ = p2pkh.Lock(changeAddress)
		tx.AddOutput(change)
	}

	return tx, nil
}

func BuildInscription(id string, amt uint64) *script.Script {
	transferJSON := fmt.Sprintf(`{"p":"bsv-20","op":"transfer","id":"%s","amt":"%d"}`, id, amt)
	lockingScript := script.NewFromBytes([]byte{})
	_ = lockingScript.AppendOpcodes(script.OpFALSE, script.OpIF)
	_ = lockingScript.AppendPushData([]byte("ord"))
	_ = lockingScript.AppendOpcodes(script.Op1)
	_ = lockingScript.AppendPushData([]byte("application/bsv-20"))
	_ = lockingScript.AppendOpcodes(script.Op0)
	_ = lockingScript.AppendPushData([]byte(transferJSON))
	_ = lockingScript.AppendOpcodes(script.OpENDIF)
	return lockingScript
}

func (p *Pow20) Lock(supply uint64) *script.Script {
	s := BuildInscription(p.Id, supply)
	s = script.NewFromBytes(append(*s, *pow20Prefix...))
	_ = s.AppendPushData([]byte(p.Symbol))
	_ = s.AppendPushData(uint64ToBytes(p.Max))
	if p.Dec <= 16 {
		_ = s.AppendOpcodes(byte(p.Dec + 0x50))
	} else {
		_ = s.AppendPushData([]byte{p.Dec})
	}
	_ = s.AppendPushData(uint64ToBytes(p.Reward))
	_ = s.AppendOpcodes(p.Difficulty + 0x50)
	s = script.NewFromBytes(append(*s, *pow20Suffix...))

	state := script.NewFromBytes([]byte{})
	_ = state.AppendOpcodes(script.OpRETURN, script.OpFALSE)
	_ = state.AppendPushData([]byte(p.Id))
	_ = state.AppendPushData(uint64ToBytes(supply))
	stateSize := uint32(len(*state) - 1)
	stateScript := binary.LittleEndian.AppendUint32(*state, stateSize)
	stateScript = append(stateScript, 0x00)

	lockingScript := make([]byte, len(*s)+len(stateScript))
	copy(lockingScript, *s)
	copy(lockingScript[len(*s):], stateScript)
	return script.NewFromBytes(lockingScript)
}

func (o *Pow20) Unlock(nonce []byte, recipient *script.Address) (*Pow20Unlocker, error) {
	unlock := &Pow20Unlocker{
		Pow20:     *o,
		Nonce:     nonce,
		Recipient: recipient,
	}
	return unlock, nil
}

func (p *Pow20Unlocker) Sign(tx *transaction.Transaction, inputIndex uint32) (*script.Script, error) {
	unlockScript := &script.Script{}

	// pow := o.Mine(o.Char)
	_ = unlockScript.AppendPushData(p.Recipient.PublicKeyHash)
	_ = unlockScript.AppendPushData([]byte(p.Nonce))
	if preimage, err := tx.CalcInputPreimage(inputIndex, sighash.All|sighash.AnyOneCanPayForkID); err != nil {
		return nil, err
	} else {
		_ = unlockScript.AppendPushData(preimage)
	}
	var change *transaction.TransactionOutput
	for _, output := range tx.Outputs {
		if output.Change {
			if change != nil {
				return nil, errors.New("multiple change outputs")
			}
			change = output
		}
	}
	if change != nil {
		_ = unlockScript.AppendPushData(uint64ToBytes(change.Satoshis))
		_ = unlockScript.AppendPushData((*change.LockingScript)[3:23])
	} else {
		_ = unlockScript.AppendOpcodes(script.Op0, script.Op0)
	}

	return unlockScript, nil
}

func (o *Pow20Unlocker) EstimateLength(tx *transaction.Transaction, inputIndex uint32) uint32 {
	noncePrefix, _ := script.PushDataPrefix(o.Nonce)
	preimage, _ := tx.CalcInputPreimage(inputIndex, sighash.AnyOneCanPayForkID|sighash.All)
	preimagePrefix, _ := script.PushDataPrefix(preimage)

	return uint32(55 + // OP_RETURN isGenesis push recipient push change sats push change pkh
		len(noncePrefix) + len(o.Nonce) + // push data ownerScript
		len(preimagePrefix) + len(preimage)) // push data preimage

}
