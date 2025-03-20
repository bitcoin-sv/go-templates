package bsocial

import (
	"github.com/bitcoin-sv/go-templates/template/bitcom"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
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

type Action struct {
	App             string  `json:"app"`
	Type            string  `json:"type"`
	Context         Context `json:"context,omitempty"`
	ContextValue    string  `json:"contextValue,omitempty"`
	Subcontext      Context `json:"subcontext,omitempty"`
	SubcontextValue string  `json:"subcontextValue,omitempty"`
}

// Post represents a new piece of content
type Post struct {
	Action
	B bitcom.B `json:"b"`
}

type Reply struct {
	Action
	B bitcom.B `json:"b"`
}

// Like represents liking a post
type Like struct {
	Action
}

// Unlike represents unliking a post
type Unlike struct {
	Action
}

// Follow represents following a user
type Follow struct {
	Action
}

// Unfollow represents unfollowing a user
type Unfollow struct {
	Action
}

// Message represents a message in a channel or to a user
type Message struct {
	Action
	B bitcom.B `json:"b"`
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
	Tags        [][]string  `json:"tags,omitempty"`
}

func DecodeTransaction(tx *transaction.Transaction) (bsocial *BSocial) {
	bsocial = &BSocial{}

	for _, output := range tx.Outputs {
		if bc := bitcom.Decode(output.LockingScript); bc != nil {
			for _, proto := range bc.Protocols {
				switch proto.Protocol {
				case bitcom.MapPrefix:
					if m := bitcom.DecodeMap(proto.Script); m != nil {
						// Check for tags in MAP data
						if m.Data["app"] == AppName && m.Data["type"] == "post" {
							// Try to extract tags if present
							if tagsField, exists := m.Data["tags"]; exists {
								// Handle different tag formats
								processTags(bsocial, tagsField)
								continue
							}
						}

						// Process by type
						switch m.Data["type"] {
						case "post":
							bsocial.Post = &Post{
								B: bitcom.B{
									MediaType: bitcom.MediaType(m.Data["mediaType"]),
									Encoding:  bitcom.Encoding(m.Data["encoding"]),
									Data:      []byte(m.Data["content"]),
								},
								Action: Action{
									Type:            "post",
									Context:         Context(m.Data["context"]),
									ContextValue:    m.Data["contextValue"],
									Subcontext:      Context(m.Data["subcontext"]),
									SubcontextValue: m.Data["subcontextValue"],
								},
							}
						case "reply":
							bsocial.Reply = &Reply{
								B: bitcom.B{
									MediaType: bitcom.MediaType(m.Data["mediaType"]),
									Encoding:  bitcom.Encoding(m.Data["encoding"]),
									Data:      []byte(m.Data["content"]),
								},
								Action: Action{
									Type:         "reply",
									Context:      Context(m.Data["context"]),
									ContextValue: m.Data["contextValue"],
								},
							}
						case "like":
							bsocial.Like = &Like{
								Action: Action{
									Type:         "like",
									Context:      ContextTx,
									ContextValue: m.Data["tx"],
								},
							}
						case "unlike":
							bsocial.Unlike = &Unlike{
								Action: Action{
									Type:         "unlike",
									Context:      ContextTx,
									ContextValue: m.Data["tx"],
								},
							}
						case "follow":
							bsocial.Follow = &Follow{
								Action: Action{
									Type:         "follow",
									Context:      ContextBapID,
									ContextValue: m.Data["bapID"],
								},
							}
						case "unfollow":
							bsocial.Unfollow = &Unfollow{
								Action: Action{
									Type:         "unfollow",
									Context:      ContextBapID,
									ContextValue: m.Data["bapID"],
								},
							}
						case "message":
							bsocial.Message = &Message{
								B: bitcom.B{
									MediaType: bitcom.MediaType(m.Data["mediaType"]),
									Encoding:  bitcom.Encoding(m.Data["encoding"]),
									Data:      []byte(m.Data["content"]),
								},
								Action: Action{
									Type:         "message",
									Context:      Context(m.Data["context"]),
									ContextValue: m.Data["contextValue"],
								},
							}
						}
					}
				case bitcom.BPrefix:
					// For B protocol handling, use the original script
					// Since this is complex, we'll skip B attachment processing for now
					// TODO: Implement proper B protocol attachment processing
				}
			}
		}
	}

	// If bsocial is empty (no fields set), return nil
	if bsocial.Post == nil && bsocial.Reply == nil && bsocial.Like == nil &&
		bsocial.Unlike == nil && bsocial.Follow == nil && bsocial.Unfollow == nil &&
		bsocial.Message == nil && bsocial.AIP == nil && len(bsocial.Attachments) == 0 &&
		len(bsocial.Tags) == 0 {
		return nil
	}

	return
}

// processTags handles different tag formats and adds them to the BSocial object
func processTags(bsocial *BSocial, tagsField interface{}) {
	// Handle string
	if tagStr, ok := tagsField.(string); ok {
		bsocial.Tags = append(bsocial.Tags, []string{tagStr})
		return
	}

	// Handle []string
	if tagSlice, ok := tagsField.([]string); ok {
		bsocial.Tags = append(bsocial.Tags, tagSlice)
		return
	}

	// Handle []interface{}
	if tagIface, ok := tagsField.([]interface{}); ok {
		var parsedTags []string
		for _, t := range tagIface {
			if ts, ok := t.(string); ok {
				parsedTags = append(parsedTags, ts)
			}
		}
		if len(parsedTags) > 0 {
			bsocial.Tags = append(bsocial.Tags, parsedTags)
		}
	}
}

// CreatePost creates a new post transaction
func CreatePost(post Post, tags []string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.BPrefix))
	s.AppendPushData(post.B.Data)
	s.AppendPushData([]byte(string(post.B.MediaType)))
	s.AppendPushData([]byte(string(post.B.Encoding)))
	if post.B.Filename != "" {
		s.AppendPushData([]byte(post.B.Filename))
	}

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
	if len(tags) > 0 {
		tagsScript := &script.Script{}
		tagsScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
		tagsScript.AppendPushData([]byte(bitcom.MapPrefix))
		tagsScript.AppendPushData([]byte("SET"))
		tagsScript.AppendPushData([]byte("app"))
		tagsScript.AppendPushData([]byte(AppName))
		tagsScript.AppendPushData([]byte("type"))
		tagsScript.AppendPushData([]byte("post"))
		tagsScript.AppendPushData([]byte("tags"))
		for _, tag := range tags {
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
func CreateReply(reply Reply, replyTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, privateKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.BPrefix))
	s.AppendPushData(reply.B.Data)
	s.AppendPushData([]byte(string(reply.B.MediaType)))
	s.AppendPushData([]byte(string(reply.B.Encoding)))
	if reply.B.Filename != "" {
		s.AppendPushData([]byte(reply.B.Filename))
	}

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
	s.AppendPushData(msg.B.Data)
	s.AppendPushData([]byte(string(msg.B.MediaType)))
	s.AppendPushData([]byte(string(msg.B.Encoding)))
	if msg.B.Filename != "" {
		s.AppendPushData([]byte(msg.B.Filename))
	}

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
