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
	// AppName is the default application name for BSocial actions
	AppName = "bsocial"
)

// Action represents a base BSocial action with common fields
type Action struct {
	App             string        `json:"app"`
	Type            ActionType    `json:"type"`
	Context         ActionContext `json:"context,omitempty"`
	ContextValue    string        `json:"contextValue,omitempty"`
	Subcontext      ActionContext `json:"subcontext,omitempty"`
	SubcontextValue string        `json:"subcontextValue,omitempty"`
}

// Ord represents an Ordinal
// Used to define collection properties on 1Sat Ordinals
type Ord struct {
	Action
}

// Claim represents a claim to something
// Used to claim OPNS handles, etc.
type Claim struct {
	Action
}

// Post represents a new piece of content
type Post struct {
	Action
	B bitcom.B `json:"b"`
}

// Reply represents a reply to an existing post
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

// BMap represents a collection of BitCom protocol data
type BMap struct {
	MAP []bitcom.Map `json:"map"`
	B   []bitcom.B   `json:"b"`
	AIP []bitcom.AIP `json:"aip,omitempty"`
}

// BSocial represents all potential BSocial actions for a transaction
type BSocial struct {
	Ord         *Ord        `json:"ord"`
	Claim       *Claim      `json:"claim"`
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

// DecodeTransaction parses a transaction and extracts BSocial protocol data
func DecodeTransaction(tx *transaction.Transaction) (bsocial *BSocial) {
	bsocial = &BSocial{}

	for _, output := range tx.Outputs {
		if output.LockingScript == nil {
			continue
		}

		if bc := bitcom.Decode(output.LockingScript); bc != nil {
			processProtocols(bc, bsocial)
		}
	}
	var trimAttachments bool
	if bsocial.Post != nil && len(bsocial.Attachments) > 0 {

		bsocial.Post.B = bsocial.Attachments[0]
		trimAttachments = true
	}

	if bsocial.Reply != nil && len(bsocial.Attachments) > 0 {
		bsocial.Reply.B = bsocial.Attachments[0]
		trimAttachments = true
	}

	if bsocial.Message != nil && len(bsocial.Attachments) > 0 {
		bsocial.Message.B = bsocial.Attachments[0]
		trimAttachments = true
	}

	if trimAttachments {
		if len(bsocial.Attachments) > 1 {
			bsocial.Attachments = bsocial.Attachments[1:]
		} else {
			bsocial.Attachments = nil
		}
	}

	// If bsocial is empty (no fields set), return nil
	if bsocial.IsEmpty() {
		return nil
	}

	return
}

// processProtocols extracts and processes BitCom protocol data
func processProtocols(bc *bitcom.Bitcom, bsocial *BSocial) {
	for _, proto := range bc.Protocols {
		switch proto.Protocol {
		case bitcom.MapPrefix:
			if m := bitcom.DecodeMap(proto.Script); m != nil {
				processMapData(m, bsocial)
			} else {
				fmt.Printf("Failed to decode MAP data: %s\n", proto.Script)
			}
		case bitcom.BPrefix:
			if b := bitcom.DecodeB(proto.Script); b != nil {
				bsocial.Attachments = append(bsocial.Attachments, *b)
			}
		default:
			// Silently ignore unknown protocols
		}
	}
}

// processMapData analyzes MAP data and populates the BSocial object
func processMapData(m *bitcom.Map, bsocial *BSocial) {
	// Check for tags in MAP data
	if m.Data["app"] == AppName && m.Data["type"] == "post" {
		// Try to extract tags if present
		if tagsField, exists := m.Data["tags"]; exists {
			processTags(bsocial, tagsField)
			return
		}
	}

	// Type-specific handlers mapped to action types
	handlers := map[ActionType]func(*bitcom.Map, *BSocial){
		TypePostReply: func(m *bitcom.Map, bs *BSocial) {
			// Check if this is a reply (has a context_tx) or a regular post
			if _, exists := m.Data["tx"]; exists {
				// This is a reply
				bs.Reply = &Reply{
					Action: createAction(TypePostReply, m),
				}
			} else {
				// This is a regular post
				bs.Post = &Post{
					Action: createAction(TypePostReply, m),
				}
			}
		},
		TypeLike: func(m *bitcom.Map, bs *BSocial) {
			bs.Like = &Like{
				Action: Action{
					Type:         TypeLike,
					Context:      ContextTx,
					ContextValue: m.Data["tx"],
				},
			}
		},
		TypeUnlike: func(m *bitcom.Map, bs *BSocial) {
			bs.Unlike = &Unlike{
				Action: Action{
					Type:         TypeUnlike,
					Context:      ContextTx,
					ContextValue: m.Data["tx"],
				},
			}
		},
		TypeFollow: func(m *bitcom.Map, bs *BSocial) {
			bs.Follow = &Follow{
				Action: Action{
					Type:         TypeFollow,
					Context:      ContextBapID,
					ContextValue: m.Data["bapID"],
				},
			}
		},
		TypeUnfollow: func(m *bitcom.Map, bs *BSocial) {
			bs.Unfollow = &Unfollow{
				Action: Action{
					Type:         TypeUnfollow,
					Context:      ContextBapID,
					ContextValue: m.Data["bapID"],
				},
			}
		},
		TypeMessage: func(m *bitcom.Map, bs *BSocial) {
			bs.Message = &Message{
				Action: createAction(TypeMessage, m),
			}
		},
	}

	// Execute the appropriate handler if one exists for this action type
	if actionType := ActionType(m.Data["type"]); actionType != "" {
		if handler, exists := handlers[actionType]; exists {
			handler(m, bsocial)
		}
	}
}

// createAction builds an Action structure from MAP data
func createAction(actionType ActionType, m *bitcom.Map) Action {
	action := Action{
		App:        m.Data["app"],
		Type:       actionType,
		Context:    ActionContext(m.Data["context"]),
		Subcontext: ActionContext(m.Data["subcontext"]),
	}

	if context, exists := m.Data["context"]; exists {
		action.ContextValue = m.Data[context]
	}
	if subcontext, exists := m.Data["subcontext"]; exists {
		action.SubcontextValue = m.Data[subcontext]
	}
	return action
}

// CreatePost creates a new post transaction
func CreatePost(post Post, attachments []bitcom.B, tags []string, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()

	// Create B protocol output first
	s := &script.Script{}
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.BPrefix)
	_ = s.AppendPushData(post.B.Data)
	_ = s.AppendPushDataString(string(post.B.MediaType))
	_ = s.AppendPushDataString(string(post.B.Encoding))
	if post.B.Filename != "" {
		_ = s.AppendPushDataString(post.B.Filename)
	}

	// Add MAP protocol
	_ = s.AppendPushDataString("|")
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(post.App)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypePostReply))

	// Add context if provided
	if post.Context != "" {
		_ = s.AppendPushDataString(string(post.Context))
		_ = s.AppendPushDataString(post.ContextValue)
	}

	// Add subcontext if provided
	if post.Subcontext != "" {
		_ = s.AppendPushDataString(string(post.Subcontext))
		_ = s.AppendPushDataString(post.SubcontextValue)
	}

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")

		// make a string from the mapScript
		data := s.String()
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
		// pubKey := identityKey.PubKey()
		// mapScript.AppendPushData(pubKey.Compressed())
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Add tags if present
	if len(tags) > 0 {
		tagsScript := &script.Script{}
		_ = tagsScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
		_ = tagsScript.AppendPushDataString(bitcom.MapPrefix)
		_ = tagsScript.AppendPushDataString("ADD")
		_ = tagsScript.AppendPushDataString("tags")
		for _, tag := range tags {
			_ = tagsScript.AppendPushDataString(tag)
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
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.BPrefix)
	_ = s.AppendPushData(reply.B.Data)
	_ = s.AppendPushDataString(string(reply.B.MediaType))
	_ = s.AppendPushDataString(string(reply.B.Encoding))
	if reply.B.Filename != "" {
		_ = s.AppendPushDataString(reply.B.Filename)
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output

	_ = s.AppendPushDataString("|")
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(AppName)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypePostReply))
	_ = s.AppendPushDataString("context")
	_ = s.AppendPushDataString("tx")
	_ = s.AppendPushDataString("tx")
	_ = s.AppendPushDataString(replyTxID)

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the mapScript
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
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

// CreateLike creates a like transaction
func CreateLike(likeTxID string, utxos []*transaction.UTXO, changeAddress *script.Address, identityKey *ec.PrivateKey) (*transaction.Transaction, error) {
	tx := transaction.NewTransaction()
	s := &script.Script{}
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(AppName)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypeLike))
	_ = s.AppendPushDataString("context")
	_ = s.AppendPushDataString(string(ContextTx))
	_ = s.AppendPushDataString(string(ContextTx))
	_ = s.AppendPushDataString(likeTxID)

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
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
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(AppName)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypeUnlike))
	_ = s.AppendPushDataString("context")
	_ = s.AppendPushDataString(string(ContextTx))
	_ = s.AppendPushDataString(string(ContextTx))
	_ = s.AppendPushDataString(unlikeTxID)

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
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
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(AppName)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypeFollow))
	_ = s.AppendPushDataString("context")
	_ = s.AppendPushDataString(string(ContextBapID))
	_ = s.AppendPushDataString(string(ContextBapID))
	_ = s.AppendPushDataString(followBapID)

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
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
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.MapPrefix)
	_ = s.AppendPushDataString("SET")
	_ = s.AppendPushDataString("app")
	_ = s.AppendPushDataString(AppName)
	_ = s.AppendPushDataString("type")
	_ = s.AppendPushDataString(string(TypeUnfollow))
	_ = s.AppendPushDataString("context")
	_ = s.AppendPushDataString(string(ContextBapID))
	_ = s.AppendPushDataString(string(ContextBapID))
	_ = s.AppendPushDataString(unfollowBapID)

	// Add AIP signature
	if identityKey != nil {
		_ = s.AppendPushDataString("|")
		_ = s.AppendPushDataString(bitcom.AIPPrefix)
		_ = s.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the script
		data := s.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = s.AppendPushDataString(sig.Signature)
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
	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = s.AppendPushDataString(bitcom.BPrefix)
	_ = s.AppendPushData(message.B.Data)
	_ = s.AppendPushDataString(string(message.B.MediaType))
	_ = s.AppendPushDataString(string(message.B.Encoding))
	if message.B.Filename != "" {
		_ = s.AppendPushDataString(message.B.Filename)
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: s,
		Satoshis:      0,
	})

	// Create MAP protocol output
	mapScript := &script.Script{}
	_ = mapScript.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	_ = mapScript.AppendPushDataString(bitcom.MapPrefix)
	_ = mapScript.AppendPushDataString("SET")
	_ = mapScript.AppendPushDataString("app")
	_ = mapScript.AppendPushDataString(AppName)
	_ = mapScript.AppendPushDataString("type")
	_ = mapScript.AppendPushDataString(string(TypeMessage))

	// Add context if provided
	if message.Context != "" {
		_ = mapScript.AppendPushDataString("context")
		_ = mapScript.AppendPushDataString(string(message.Context))
		_ = mapScript.AppendPushDataString(string(message.Context))
		_ = mapScript.AppendPushDataString(message.ContextValue)
	}

	// Add AIP signature
	if identityKey != nil {
		_ = mapScript.AppendPushDataString("|")
		_ = mapScript.AppendPushDataString(bitcom.AIPPrefix)
		_ = mapScript.AppendPushDataString("BITCOIN_ECDSA")

		// make a string from the mapScript
		data := mapScript.String()
		sig, err := aip.Sign(identityKey, aip.BitcoinECDSA, data)
		if err != nil {
			return nil, err
		}
		_ = mapScript.AppendPushDataString(sig.Signature)
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
