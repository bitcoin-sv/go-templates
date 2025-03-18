package bsocial

import (
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/bsv-blockchain/go-templates/template/bitcom"
)

const (
	// Protocol prefixes
	// bitcom.MapPrefix = "1PuQa7K62MiKCtssSLKy1kh56WWU7MtUR5"

	AppName = "bsocial"
)

// Encoding types
type Encoding string

const (
	EncodingUTF8   Encoding = "utf-8"
	EncodingBase64 Encoding = "base64"
	EncodingHex    Encoding = "hex"
)

// Context types
type Context string

const (
	ContextTx       Context = "tx"
	ContextChannel  Context = "channel"
	ContextBapID    Context = "bapID"
	ContextProvider Context = "provider"
	ContextVideoID  Context = "videoID"
)

// Post represents a new piece of content
type Post struct {
	MediaType       bitcom.MediaType `json:"mediaType"`
	Encoding        Encoding         `json:"encoding"`
	Content         string           `json:"content"`
	Context         Context          `json:"context,omitempty"`
	ContextValue    string           `json:"contextValue,omitempty"`
	Subcontext      Context          `json:"subcontext,omitempty"`
	SubcontextValue string           `json:"subcontextValue,omitempty"`
	// Tags            []string         `json:"tags,omitempty"`
}

// Message represents a message in a channel or to a user
type Message struct {
	MediaType    bitcom.MediaType `json:"mediaType"`
	Encoding     bitcom.Encoding  `json:"encoding"`
	Content      string           `json:"content"`
	Context      Context          `json:"context"`
	ContextValue string           `json:"contextValue"`
}

type BMap struct {
	MAP []bitcom.Map `json:"map"`
	B   []bitcom.B   `json:"b"`
	AIP []bitcom.AIP `json:"aip,omitempty"`
}

type BSocial struct {
	Post        *Post       `json:"post"`
	Reply       *Reply      `json:"reply"`
	Like        *Like       `json:"like"`
	Unlike      *Unlike     `json:"unlike"`
	Follow      *Follow     `json:"follow"`
	Unfollow    *Unfollow   `json:"unfollow"`
	Message     *Message    `json:"message"`
	AIP         *bitcom.AIP `json:"aip"`
	Attachments []bitcom.B  `json:"attachments,omitempty"`
}

func DecodeTransaction(tx *transaction.Transaction) (bsocial *BSocial) {
	for _, output := range tx.Outputs {
		if bc := bitcom.Decode(output.LockingScript); bc != nil {
			for _, proto := range bc.Protocols {
				switch proto.Protocol {
				case bitcom.MapPrefix:
					if m := bitcom.DecodeMap(proto.Script); m != nil {
						switch m.Data["type"] {
						case "post":
							bsocial = &BSocial{
								Post: &Post{
									MediaType:       bitcom.MediaType(m.Data["mediaType"]),
									Encoding:        Encoding(m.Data["encoding"]),
									Content:         m.Data["content"],
									Context:         Context(m.Data["context"]),
									ContextValue:    m.Data["contextValue"],
									Subcontext:      Context(m.Data["subcontext"]),
									SubcontextValue: m.Data["subcontextValue"],
									// Tags:            m.Data["tags"],
								},
							}
						case "reply":
						case "like":
						case "unlike":
						case "follow":
						case "unfollow":
						case "message":
						}
					}
				}
			}
			// if bsocial == nil {
			// 	if bsocial = Decode(output.LockingScript); bsocial != nil {
			// 		continue
			// 	}
			// } else if bs := bitcom.DecodeB(output.LockingScript); len(bs) > 0 {
			// 	bsocial.Attachments = append(bsocial.Attachments, bs...)
			// }
		}
	}
	return
}

// // Decode
// func Decode(scr *script.Script) (bsocial *BSocial) {
// 	bc := bitcom.Decode(scr)

// 	bmap := BMap{}

// 	// Decode MAP protocol
// 	maps := bitcom.DecodeMap(bc)
// 	for _, m := range maps {
// 		switch m.Data["type"] {

// 	}

// 	// Decode B protocol
// 	bs := bitcom.DecodeB(bc)
// 	for _, b := range bs {
// 		bmap.B = append(bmap.B, *b)
// 	}

// 	// Decode AIP protocol
// 	aips := bitcom.DecodeAIP(bc)
// 	for _, aip := range aips {
// 		bmap.AIP = append(bmap.AIP, *aip)
// 	}

// 	return &bmap
// }

// CreatePost creates a new post transaction
func CreatePost(post Post, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte("B"))
	s.AppendPushData([]byte(post.Content))
	s.AppendPushData([]byte(string(post.MediaType)))
	s.AppendPushData([]byte(string(post.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("post"))

	// Add context if provided
	if post.Context != "" {
		mapScript.AppendPushData([]byte("context_" + string(post.Context)))
		mapScript.AppendPushData([]byte(post.ContextValue))
	}

	// Add subcontext if provided
	if post.Subcontext != "" {
		mapScript.AppendPushData([]byte("subcontext_" + string(post.Subcontext)))
		mapScript.AppendPushData([]byte(post.SubcontextValue))
	}

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	// Add tags if present
	if len(post.Tags) > 0 {
		tagsScript := &script.Script{}
		tagsScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
		tagsScript.AppendPushData([]byte(bitcom.MapPrefix))
		tagsScript.AppendPushData([]byte("SET"))
		tagsScript.AppendPushData([]byte("app"))
		tagsScript.AppendPushData([]byte(AppName))
		tagsScript.AppendPushData([]byte("type"))
		tagsScript.AppendPushData([]byte("post"))
		tagsScript.AppendPushData([]byte("tags"))
		for _, tag := range post.Tags {
			tagsScript.AppendPushData([]byte(tag))
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: tagsScript,
			Satoshis:      0,
		})
	}

	return tx, nil
}

// CreateReply creates a reply to an existing post
func CreateReply(reply Post, replyTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.BPrefix))
	s.AppendPushData([]byte(reply.Content))
	s.AppendPushData([]byte(string(reply.MediaType)))
	s.AppendPushData([]byte(string(reply.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("post"))
	mapScript.AppendPushData([]byte("context_tx"))
	mapScript.AppendPushData([]byte(replyTxID))

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	return tx, nil
}

// CreateLike creates a like transaction
func CreateLike(likeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("like"))
	s.AppendPushData([]byte("tx"))
	s.AppendPushData([]byte(likeTxID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(bitcom.AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateUnlike creates an unlike transaction
func CreateUnlike(unlikeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("unlike"))
	s.AppendPushData([]byte("tx"))
	s.AppendPushData([]byte(unlikeTxID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(bitcom.AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateFollow creates a follow transaction
func CreateFollow(followBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("follow"))
	s.AppendPushData([]byte("bapID"))
	s.AppendPushData([]byte(followBapID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(bitcom.AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateUnfollow creates an unfollow transaction
func CreateUnfollow(unfollowBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte("unfollow"))
	s.AppendPushData([]byte("bapID"))
	s.AppendPushData([]byte(unfollowBapID))
	s.AppendPushData([]byte("|"))
	s.AppendPushData([]byte(bitcom.AIPPrefix))
	s.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	s.AppendPushData(pubKey.Compressed())

	// TODO: Add proper signature calculation

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// TODO: Add proper UTXO handling and change address

	return tx, nil
}

// CreateMessage creates a new message transaction
func CreateMessage(msg Message, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.BPrefix))
	s.AppendPushData([]byte(msg.Content))
	s.AppendPushData([]byte(string(msg.MediaType)))
	s.AppendPushData([]byte(string(msg.Encoding)))
	s.AppendPushData([]byte("UTF8"))

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	mapScript.AppendPushData([]byte(bitcom.MapPrefix))
	mapScript.AppendPushData([]byte("SET"))
	mapScript.AppendPushData([]byte("app"))
	mapScript.AppendPushData([]byte(AppName))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte("message"))
	mapScript.AppendPushData([]byte("context_" + string(msg.Context)))
	mapScript.AppendPushData([]byte(msg.ContextValue))

	// Add AIP signature
	mapScript.AppendPushData([]byte("|"))
	mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
	mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))
	pubKey := privateKey.PubKey()
	mapScript.AppendPushData(pubKey.Compressed())

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	return tx, nil
}
