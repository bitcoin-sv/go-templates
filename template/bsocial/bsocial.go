package bsocial

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bitcoin-sv/go-templates/template/p2pkh"
	"github.com/bitcoinschema/go-aip"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

const (
	AppName = "bsocial"
)

type Action struct {
	App             string      `json:"app"`
	Type            BSocialType `json:"type"`
	Context         Context     `json:"context,omitempty"`
	ContextValue    string      `json:"contextValue,omitempty"`
	Subcontext      Context     `json:"subcontext,omitempty"`
	SubcontextValue string      `json:"subcontextValue,omitempty"`
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

// DecodeTransaction decodes a transaction and returns a BSocial object
func DecodeTransaction(tx *transaction.Transaction) (bsocial *BSocial) {
	bsocial = &BSocial{}

	// Print debug information
	fmt.Printf("Decoding transaction with %d outputs\n", len(tx.Outputs))

	for i, output := range tx.Outputs {
		fmt.Printf("Processing output #%d\n", i)
		if output.LockingScript == nil {
			fmt.Printf("Output #%d has nil LockingScript\n", i)
			continue
		}

		if bc := bitcom.Decode(output.LockingScript); bc != nil {
			fmt.Printf("Output #%d contains BitCom data with %d protocols\n", i, len(bc.Protocols))
			processProtocols(bc, output.LockingScript, bsocial)
		}
	}

	// If bsocial is empty (no fields set), return nil
	if bsocial.IsEmpty() {
		fmt.Println("BSocial object is empty, returning nil")
		return nil
	}

	fmt.Println("BSocial object created successfully")
	return
}

// processProtocols processes all bitcom protocols in the script
func processProtocols(bc *bitcom.Bitcom, script *script.Script, bsocial *BSocial) {
	fmt.Printf("Processing %d protocols\n", len(bc.Protocols))

	for i, proto := range bc.Protocols {
		fmt.Printf("Protocol #%d: %s\n", i, proto.Protocol)

		switch proto.Protocol {
		case bitcom.MapPrefix:
			fmt.Printf("Found MAP protocol\n")
			if m := bitcom.DecodeMap(proto.Script); m != nil {
				fmt.Printf("MAP data decoded: %+v\n", m.Data)
				processMapData(m, bsocial)
			} else {
				fmt.Printf("Failed to decode MAP data\n")
			}
		case bitcom.BPrefix:
			fmt.Printf("Found B protocol\n")
			if b := bitcom.DecodeB(script); b != nil {
				fmt.Printf("B data decoded: MediaType=%s, Encoding=%s, DataLength=%d\n",
					b.MediaType, b.Encoding, len(b.Data))
				bsocial.Attachments = append(bsocial.Attachments, *b)
			} else {
				fmt.Printf("Failed to decode B data\n")
			}
		default:
			fmt.Printf("Unknown protocol: %s\n", proto.Protocol)
		}
	}
}

// processMapData processes MAP protocol data based on action type
func processMapData(m *bitcom.Map, bsocial *BSocial) {
	fmt.Printf("Processing MAP data: app=%s, type=%s\n", m.Data["app"], m.Data["type"])

	// Check for tags in MAP data
	if m.Data["app"] == AppName && m.Data["type"] == "post" {
		// Try to extract tags if present
		if tagsField, exists := m.Data["tags"]; exists {
			fmt.Printf("Found tags field: %v\n", tagsField)
			processTags(bsocial, tagsField)
			return
		}
	}

	// Type-specific handlers mapped to action types
	handlers := map[BSocialType]func(*bitcom.Map, *BSocial){
		TypePostReply: func(m *bitcom.Map, bs *BSocial) {
			// Check if this is a reply (has a context_tx) or a regular post
			if _, exists := m.Data["tx"]; exists {
				// This is a reply
				fmt.Printf("Creating a Reply object\n")
				bs.Reply = &Reply{
					B:      createB(m),
					Action: createAction(TypePostReply, m),
				}
			} else {
				// This is a regular post
				fmt.Printf("Creating a Post object\n")
				bs.Post = &Post{
					B:      createB(m),
					Action: createAction(TypePostReply, m),
				}
			}
		},
		TypeLike: func(m *bitcom.Map, bs *BSocial) {
			fmt.Printf("Creating a Like object with tx=%s\n", m.Data["tx"])
			bs.Like = &Like{
				Action: Action{
					Type:         TypeLike,
					Context:      ContextTx,
					ContextValue: m.Data["tx"],
				},
			}
		},
		TypeUnlike: func(m *bitcom.Map, bs *BSocial) {
			fmt.Printf("Creating an Unlike object with tx=%s\n", m.Data["tx"])
			bs.Unlike = &Unlike{
				Action: Action{
					Type:         TypeUnlike,
					Context:      ContextTx,
					ContextValue: m.Data["tx"],
				},
			}
		},
		TypeFollow: func(m *bitcom.Map, bs *BSocial) {
			fmt.Printf("Creating a Follow object with bapID=%s\n", m.Data["bapID"])
			bs.Follow = &Follow{
				Action: Action{
					Type:         TypeFollow,
					Context:      ContextBapID,
					ContextValue: m.Data["bapID"],
				},
			}
		},
		TypeUnfollow: func(m *bitcom.Map, bs *BSocial) {
			fmt.Printf("Creating an Unfollow object with bapID=%s\n", m.Data["bapID"])
			bs.Unfollow = &Unfollow{
				Action: Action{
					Type:         TypeUnfollow,
					Context:      ContextBapID,
					ContextValue: m.Data["bapID"],
				},
			}
		},
		TypeMessage: func(m *bitcom.Map, bs *BSocial) {
			fmt.Printf("Creating a Message object\n")
			bs.Message = &Message{
				B:      createB(m),
				Action: createAction(TypeMessage, m),
			}
		},
	}

	// Execute the appropriate handler if one exists for this action type
	if actionType := BSocialType(m.Data["type"]); actionType != "" {
		fmt.Printf("Looking for handler for action type: %s\n", actionType)
		if handler, exists := handlers[actionType]; exists {
			fmt.Printf("Handler found for action type: %s\n", actionType)
			handler(m, bsocial)
		} else {
			fmt.Printf("No handler found for action type: %s\n", actionType)
		}
	} else {
		fmt.Printf("No action type found in MAP data\n")
	}
}

// Helper functions to create common structures
func createB(m *bitcom.Map) bitcom.B {
	return bitcom.B{
		MediaType: bitcom.MediaType(m.Data["mediaType"]),
		Encoding:  bitcom.Encoding(m.Data["encoding"]),
		Data:      []byte(m.Data["content"]),
	}
}

func createAction(actionType BSocialType, m *bitcom.Map) Action {
	return Action{
		Type:            actionType,
		Context:         Context(m.Data["context"]),
		ContextValue:    m.Data["contextValue"],
		Subcontext:      Context(m.Data["subcontext"]),
		SubcontextValue: m.Data["subcontextValue"],
	}
}

// CreatePost creates a new post transaction
func CreatePost(post Post, attachments []bitcom.B, tags []string, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
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
	mapScript.AppendPushData([]byte(post.App))
	mapScript.AppendPushData([]byte("type"))
	mapScript.AppendPushData([]byte(string(TypePostReply)))

	// Add context if provided
	if post.Context != "" {
		mapScript.AppendPushData([]byte(string(post.Context)))
		mapScript.AppendPushData([]byte(post.ContextValue))
	}

	// Add subcontext if provided
	if post.Subcontext != "" {
		mapScript.AppendPushData([]byte(string(post.Subcontext)))
		mapScript.AppendPushData([]byte(post.SubcontextValue))
	}

	// Add AIP signature
	if identityKey != nil {
		mapScript.AppendPushData([]byte("|"))
		mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
		mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the mapScript
		data := mapScript.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		mapScript.AppendPushData([]byte(sig.Signature))
		// pubKey := identityKey.PubKey()
		// mapScript.AppendPushData(pubKey.Compressed())
	}

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
		tagsScript.AppendPushData([]byte(string(TypePostReply)))
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
func CreateReply(reply Reply, replyTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
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
	mapScript.AppendPushData([]byte(string(TypePostReply)))
	mapScript.AppendPushData([]byte("context"))
	mapScript.AppendPushData([]byte("tx"))
	mapScript.AppendPushData([]byte("tx"))
	mapScript.AppendPushData([]byte(replyTxID))

	// Add AIP signature
	if identityKey != nil {
		mapScript.AppendPushData([]byte("|"))
		mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
		mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the mapScript
		data := mapScript.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		mapScript.AppendPushData([]byte(sig.Signature))
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// CreateLike creates a like transaction
func CreateLike(likeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte(string(TypeLike)))
	s.AppendPushData([]byte("context"))
	s.AppendPushData([]byte(string(ContextTx)))
	s.AppendPushData([]byte(string(ContextTx)))
	s.AppendPushData([]byte(likeTxID))

	// Add AIP signature
	if identityKey != nil {
		s.AppendPushData([]byte("|"))
		s.AppendPushData([]byte(bitcom.AIPPrefix))
		s.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		s.AppendPushData([]byte(sig.Signature))
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// CreateUnlike creates an unlike transaction
func CreateUnlike(unlikeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte(string(TypeUnlike)))
	s.AppendPushData([]byte("context"))
	s.AppendPushData([]byte(string(ContextTx)))
	s.AppendPushData([]byte(string(ContextTx)))
	s.AppendPushData([]byte(unlikeTxID))

	// Add AIP signature
	if identityKey != nil {
		s.AppendPushData([]byte("|"))
		s.AppendPushData([]byte(bitcom.AIPPrefix))
		s.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		s.AppendPushData([]byte(sig.Signature))
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// CreateFollow creates a follow transaction
func CreateFollow(followBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte(string(TypeFollow)))
	s.AppendPushData([]byte("context"))
	s.AppendPushData([]byte(string(ContextBapID)))
	s.AppendPushData([]byte(string(ContextBapID)))
	s.AppendPushData([]byte(followBapID))

	// Add AIP signature
	if identityKey != nil {
		s.AppendPushData([]byte("|"))
		s.AppendPushData([]byte(bitcom.AIPPrefix))
		s.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		s.AppendPushData([]byte(sig.Signature))
	}

	// Add action output
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// CreateUnfollow creates an unfollow transaction
func CreateUnfollow(unfollowBapID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.MapPrefix))
	s.AppendPushData([]byte("SET"))
	s.AppendPushData([]byte("app"))
	s.AppendPushData([]byte(AppName))
	s.AppendPushData([]byte("type"))
	s.AppendPushData([]byte(string(TypeUnfollow)))
	s.AppendPushData([]byte("context"))
	s.AppendPushData([]byte(string(ContextBapID)))
	s.AppendPushData([]byte(string(ContextBapID)))
	s.AppendPushData([]byte(unfollowBapID))

	// Add AIP signature
	if identityKey != nil {
		s.AppendPushData([]byte("|"))
		s.AppendPushData([]byte(bitcom.AIPPrefix))
		s.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		s.AppendPushData([]byte(sig.Signature))
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// CreateMessage creates a new message transaction
func CreateMessage(message Message, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	s.AppendPushData([]byte(bitcom.BPrefix))
	s.AppendPushData(message.B.Data)
	s.AppendPushData([]byte(string(message.B.MediaType)))
	s.AppendPushData([]byte(string(message.B.Encoding)))
	if message.B.Filename != "" {
		s.AppendPushData([]byte(message.B.Filename))
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
	mapScript.AppendPushData([]byte(string(TypeMessage)))

	// Add context if provided
	if message.Context != "" {
		mapScript.AppendPushData([]byte("context"))
		mapScript.AppendPushData([]byte(string(message.Context)))
		mapScript.AppendPushData([]byte(string(message.Context)))
		mapScript.AppendPushData([]byte(message.ContextValue))
	}

	// Add AIP signature
	if identityKey != nil {
		mapScript.AppendPushData([]byte("|"))
		mapScript.AppendPushData([]byte(bitcom.AIPPrefix))
		mapScript.AppendPushData([]byte("BITCOIN_ECDSA"))

		// make a string from the mapScript
		data := mapScript.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		mapScript.AppendPushData([]byte(sig.Signature))
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: mapScript,
		Satoshis:      0,
	})

	// Add change output if changeAddress is provided
	if changeAddress != nil {
		changeScript, err := p2pkh.Lock(changeAddress)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	return tx, nil
}

// processTags handles different tag formats and adds them to the BSocial object
func processTags(bsocial *BSocial, tagsField any) {
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

	// Handle []any
	if tagIface, ok := tagsField.([]any); ok {
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
