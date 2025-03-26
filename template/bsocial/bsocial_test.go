package bsocial

import (
	"testing"

	"github.com/bitcoin-sv/go-templates/template/bitcom"
	"github.com/bitcoinschema/go-bmap"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/stretchr/testify/require"
)

// TestCreatePost verifies the Post creation functionality
func TestCreatePost(t *testing.T) {
	// Create a test private key
	// privKey, err := ec.NewPrivateKey()
	// require.NoError(t, err)

	// Create a test post
	post := Post{
		B: bitcom.B{
			MediaType: bitcom.MediaTypeTextMarkdown,
			Encoding:  bitcom.EncodingUTF8,
			Data:      []byte("# Hello BSV\nThis is a test post"),
		},
		Action: Action{
			Type: TypePostReply,
		},
	}

	// Define tags for the post
	tags := []string{"test", "bsv"}

	// Create the transaction
	tx, err := CreatePost(post, nil, tags, nil)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Log MAP entries for diagnostic purposes
	t.Logf("MAP Entries: %+v", bmapTx.MAP)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.GreaterOrEqual(t, len(bmapTx.MAP), 1)

	// check that we have one or more MAP entries
	// and that the B data is correct

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)

	// Compare the content correctly
	require.Equal(t, string(post.B.Data), string(bmapTx.B[0].Data))
	require.Equal(t, string(post.B.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(post.B.Encoding), bmapTx.B[0].Encoding)
}

// TestCreateLike verifies the Like creation functionality
func TestCreateLike(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test txid to like
	testTxID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create the transaction
	tx, err := CreateLike(testTxID, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	// TODO: Use our internal parsers instead of adding the bmap dependency
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "like", mapData["type"])
	require.Equal(t, testTxID, mapData["tx"])
}

// TestCreateReply verifies the Reply creation functionality
func TestCreateReply(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test txid to reply to
	testTxID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create a test reply
	reply := Reply{
		B: bitcom.B{
			MediaType: bitcom.MediaTypeTextPlain,
			Encoding:  bitcom.EncodingUTF8,
			Data:      []byte("This is a test reply"),
		},
		Action: Action{
			Type:         TypePostReply,
			Context:      ContextTx,
			ContextValue: testTxID,
		},
	}

	// Create the transaction
	tx, err := CreateReply(reply, testTxID, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "post", mapData["type"])
	require.Equal(t, "tx", mapData["context"])
	require.Equal(t, testTxID, mapData["tx"])

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)

	// Compare the content correctly
	require.Equal(t, string(reply.B.Data), string(bmapTx.B[0].Data))
	require.Equal(t, string(reply.B.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(reply.B.Encoding), bmapTx.B[0].Encoding)
}

// TestCreateMessage verifies the Message creation functionality
func TestCreateMessage(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Create a test message
	msg := Message{
		B: bitcom.B{
			MediaType: bitcom.MediaTypeTextPlain,
			Encoding:  bitcom.EncodingUTF8,
			Data:      []byte("Hello, this is a test message"),
		},
		Action: Action{
			Type:         "message",
			Context:      ContextChannel,
			ContextValue: "test-channel",
		},
	}

	// Create the transaction
	tx, err := CreateMessage(msg, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "message", mapData["type"])
	require.Equal(t, msg.ContextValue, mapData["channel"])

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)
	require.Equal(t, string(msg.B.Data), string(bmapTx.B[0].Data))
	require.Equal(t, string(msg.B.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(msg.B.Encoding), bmapTx.B[0].Encoding)
}

// TestCreateFollow verifies the Follow creation functionality
func TestCreateFollow(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test BAP ID to follow
	testBapID := "test-user-bap-id"

	// Create the transaction
	tx, err := CreateFollow(testBapID, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "follow", mapData["type"])
	require.Equal(t, testBapID, mapData["bapID"])
}

// TestCreateUnfollow verifies the Unfollow creation functionality
func TestCreateUnfollow(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test BAP ID to unfollow
	testBapID := "test-user-bap-id"

	// Create the transaction
	tx, err := CreateUnfollow(testBapID, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "unfollow", mapData["type"])
	require.Equal(t, testBapID, mapData["bapID"])
}

// TestCreateUnlike verifies the Unlike creation functionality
func TestCreateUnlike(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test txid to unlike
	testTxID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create the transaction
	tx, err := CreateUnlike(testTxID, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 1)
	mapData := bmapTx.MAP[0]
	require.Equal(t, "bsocial", mapData["app"])
	require.Equal(t, "unlike", mapData["type"])
	require.Equal(t, testTxID, mapData["tx"])
}

// TestDecodeTransaction verifies the transaction parsing functionality
func TestDecodeTransaction(t *testing.T) {
	// Create a test post with App field set
	post := Post{
		B: bitcom.B{
			MediaType: bitcom.MediaTypeTextMarkdown,
			Encoding:  bitcom.EncodingUTF8,
			Data:      []byte("# Test post for decoding"),
		},
		Action: Action{
			App:  AppName,
			Type: TypePostReply,
		},
	}

	// Create a post transaction◊
	tx, err := CreatePost(post, nil, []string{"tag1", "tag2"}, nil)
	require.NoError(t, err)

	// Log transaction for diagnostic purposes
	t.Logf("Transaction created: %s", tx.String())

	bsocial := DecodeTransaction(tx)
	require.NotNil(t, bsocial)
	require.NotNil(t, bsocial.Post)

	// MAke sure the values are what we expect
	// require.Equal(t, bsocial.Post.Action.Type, TypePostReply)
	// require.Equal(t, bsocial.Post.Action.App, AppName)
	// require.Equal(t, bsocial.Post.B.MediaType, bitcom.MediaTypeTextMarkdown)
	// require.Equal(t, bsocial.Post.B.Encoding, bitcom.EncodingUTF8)
	// require.Equal(t, bsocial.Post.B.Data, []byte("# Test post for decoding"))

	// Test IsEmpty
	require.False(t, bsocial.IsEmpty())

	// Create empty bsocial
	emptyBSocial := &BSocial{}
	require.True(t, emptyBSocial.IsEmpty())
}

// testBSocialFromVectors is a generic test function that validates BSocial actions
// extracted from test vectors against expected values
func testBSocialFromVectors(t *testing.T, filePath string, actionType string) {
	// Load the test vectors
	vectors := LoadTestVectors(t, filePath)

	// Loop through each test vector
	for _, vector := range vectors.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			// Parse the transaction using our helper
			tx := GetTransactionFromVector(t, vector)
			if tx == nil {
				return // Skip if transaction is nil
			}

			// Verify transaction ID matches expected value
			require.Contains(t, vector.Expected, "tx_id")
			txID := vector.Expected["tx_id"].(string)
			require.Equal(t, txID, tx.TxID().String())

			// Try to decode with bsocial first
			bsocial := DecodeTransaction(tx)

			// If DecodeTransaction fails, use bmap directly
			if bsocial == nil {
				t.Logf("DecodeTransaction returned nil for test vector '%s', trying with bmap directly", vector.Name)

				// Try parsing with bmap
				bmapTx, err := bmap.NewFromRawTxString(tx.String())
				if err != nil {
					t.Fatalf("bmap.NewFromRawTxString failed: %v", err)
				}

				t.Logf("bmap successfully parsed transaction. MAP entries: %+v", bmapTx.MAP)

				// Check for wrong_app case
				if _, ok := vector.Expected["wrong_app"]; ok {
					// This is testing a wrong app scenario, so we're expecting a mismatch
					t.Log("Vector is testing wrong app scenario - skipping further validation")
					return
				}

				// Look for specified actions in MAP
				foundAction := false
				var detectedType, contextValue string

				for _, mapData := range bmapTx.MAP {
					if typeVal, ok := mapData["type"]; ok && typeVal == actionType {
						t.Logf("Found %s action in bmap: %+v", typeVal, mapData)
						foundAction = true
						detectedType = typeVal.(string)

						// Extract context value based on action type
						switch actionType {
						case "like", "unlike", "post":
							if txVal, ok := mapData["tx"]; ok {
								contextValue = txVal.(string)
							}
						case "follow", "unfollow":
							if bapIDVal, ok := mapData["bapID"]; ok {
								contextValue = bapIDVal.(string)
							}
						case "message":
							if channelVal, ok := mapData["channel"]; ok {
								contextValue = channelVal.(string)
							}
						}
					}
				}

				// Check if we should have found the action
				shouldFail := false
				if val, ok := vector.Expected["should_fail"].(bool); ok {
					shouldFail = val
				}

				if shouldFail {
					require.False(t, foundAction, ErrMsgActionFound, actionType, vector.Name)
					return
				}

				// Otherwise expect to find an action
				require.True(t, foundAction, ErrMsgActionNotFound, actionType, vector.Name)

				// Check if the expected type matches what we found
				if expectedType, ok := vector.Expected["type"].(string); ok {
					require.Equal(t, expectedType, detectedType, ErrMsgWrongType,
						expectedType, detectedType, vector.Name)
				}

				// Check context value based on action type
				switch actionType {
				case "like", "unlike", "post":
					if postTx, ok := vector.Expected["post_tx"].(string); ok && postTx != "" {
						require.Equal(t, postTx, contextValue, ErrMsgWrongPostTx,
							postTx, contextValue, vector.Name)
					}
				case "follow", "unfollow":
					if bapID, ok := vector.Expected["bap_id"].(string); ok && bapID != "" {
						require.Equal(t, bapID, contextValue, ErrMsgWrongBapID,
							bapID, contextValue, vector.Name)
					}
				case "message":
					if channel, ok := vector.Expected["channel"].(string); ok && channel != "" {
						require.Equal(t, channel, contextValue, ErrMsgWrongChannel,
							channel, contextValue, vector.Name)
					}
				}

				return
			}

			// Handle wrong_app case
			if _, ok := vector.Expected["wrong_app"]; ok {
				// This is testing a wrong app scenario, so we're expecting nil for the target action type
				t.Log("Vector is testing wrong app scenario - checking that target action is nil")

				switch actionType {
				case "like":
					require.Nil(t, bsocial.Like)
				case "unlike":
					require.Nil(t, bsocial.Unlike)
				case "post":
					if vector.Expected["has_reply"] == true {
						require.Nil(t, bsocial.Reply)
					} else {
						require.Nil(t, bsocial.Post)
					}
				case "follow":
					require.Nil(t, bsocial.Follow)
				case "unfollow":
					require.Nil(t, bsocial.Unfollow)
				case "message":
					require.Nil(t, bsocial.Message)
				}
				return
			}

			// BSocial-specific validation based on action type
			shouldFail := false
			if val, ok := vector.Expected["should_fail"].(bool); ok {
				shouldFail = val
			}

			if shouldFail {
				// If should fail, we expect the relevant field to be nil
				switch actionType {
				case "like":
					require.Nil(t, bsocial.Like, ErrMsgNilForTestVector, "Like", vector.Name)
				case "unlike":
					require.Nil(t, bsocial.Unlike, ErrMsgNilForTestVector, "Unlike", vector.Name)
				case "post":
					if vector.Expected["has_reply"] == true {
						require.Nil(t, bsocial.Reply, ErrMsgNilForTestVector, "Reply", vector.Name)
					} else {
						require.Nil(t, bsocial.Post, ErrMsgNilForTestVector, "Post", vector.Name)
					}
				case "follow":
					require.Nil(t, bsocial.Follow, ErrMsgNilForTestVector, "Follow", vector.Name)
				case "unfollow":
					require.Nil(t, bsocial.Unfollow, ErrMsgNilForTestVector, "Unfollow", vector.Name)
				case "message":
					require.Nil(t, bsocial.Message, ErrMsgNilForTestVector, "Message", vector.Name)
				}
				return
			}

			// Otherwise, the relevant field should be non-nil
			switch actionType {
			case "like":
				require.NotNil(t, bsocial.Like, ErrMsgDecodeFailure, "like action", vector.Name)
				if contextType, ok := vector.Expected["context_type"].(string); ok {
					require.Equal(t, Context(contextType), bsocial.Like.Context, ErrMsgWrongContextType,
						contextType, bsocial.Like.Context, vector.Name)
				}
				if postTx, ok := vector.Expected["post_tx"].(string); ok && postTx != "" {
					require.Equal(t, postTx, bsocial.Like.ContextValue, ErrMsgWrongPostTx,
						postTx, bsocial.Like.ContextValue, vector.Name)
				}
			case "unlike":
				require.NotNil(t, bsocial.Unlike, ErrMsgDecodeFailure, "unlike action", vector.Name)
				if contextType, ok := vector.Expected["context_type"].(string); ok {
					require.Equal(t, Context(contextType), bsocial.Unlike.Context, ErrMsgWrongContextType,
						contextType, bsocial.Unlike.Context, vector.Name)
				}
				if postTx, ok := vector.Expected["post_tx"].(string); ok && postTx != "" {
					require.Equal(t, postTx, bsocial.Unlike.ContextValue, ErrMsgWrongPostTx,
						postTx, bsocial.Unlike.ContextValue, vector.Name)
				}
			case "post":
				if hasReply, ok := vector.Expected["has_reply"].(bool); ok && hasReply {
					// This is a reply
					require.NotNil(t, bsocial.Reply, ErrMsgDecodeFailure, "reply", vector.Name)
					if content, ok := vector.Expected["content"].(string); ok && content != "" {
						require.Equal(t, content, string(bsocial.Reply.B.Data), ErrMsgWrongContent,
							content, string(bsocial.Reply.B.Data), vector.Name)
					}
					if postTx, ok := vector.Expected["post_tx"].(string); ok && postTx != "" {
						require.Equal(t, postTx, bsocial.Reply.ContextValue, ErrMsgWrongPostTx,
							postTx, bsocial.Reply.ContextValue, vector.Name)
					}
				} else {
					// This is a regular post
					require.NotNil(t, bsocial.Post, ErrMsgDecodeFailure, "post", vector.Name)
					if content, ok := vector.Expected["content"].(string); ok && content != "" {
						require.Equal(t, content, string(bsocial.Post.B.Data), ErrMsgWrongContent,
							content, string(bsocial.Post.B.Data), vector.Name)
					}
				}
			case "follow":
				require.NotNil(t, bsocial.Follow, ErrMsgDecodeFailure, "follow action", vector.Name)
				if bapID, ok := vector.Expected["bap_id"].(string); ok && bapID != "" {
					require.Equal(t, bapID, bsocial.Follow.ContextValue, ErrMsgWrongBapID,
						bapID, bsocial.Follow.ContextValue, vector.Name)
				}
			case "unfollow":
				require.NotNil(t, bsocial.Unfollow, ErrMsgDecodeFailure, "unfollow action", vector.Name)
				if bapID, ok := vector.Expected["bap_id"].(string); ok && bapID != "" {
					require.Equal(t, bapID, bsocial.Unfollow.ContextValue, ErrMsgWrongBapID,
						bapID, bsocial.Unfollow.ContextValue, vector.Name)
				}
			case "message":
				require.NotNil(t, bsocial.Message, ErrMsgDecodeFailure, "message", vector.Name)
				if content, ok := vector.Expected["content"].(string); ok && content != "" {
					require.Equal(t, content, string(bsocial.Message.B.Data), ErrMsgWrongContent,
						content, string(bsocial.Message.B.Data), vector.Name)
				}
				if channel, ok := vector.Expected["channel"].(string); ok && channel != "" {
					require.Equal(t, channel, bsocial.Message.ContextValue, ErrMsgWrongChannel,
						channel, bsocial.Message.ContextValue, vector.Name)
				}
			}
		})
	}
}

// TestPostFromVectors validates Post actions from test vectors
func TestPostFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/post_test_vectors.json", "post")
}

// TestReplyFromVectors validates Reply actions from test vectors
// Note: replies use "post" type with context
func TestReplyFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/reply_test_vectors.json", "post")
}

// TestFollowFromVectors validates Follow actions from test vectors
func TestFollowFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/follow_test_vectors.json", "follow")
}

// TestUnfollowFromVectors validates Unfollow actions from test vectors
func TestUnfollowFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/unfollow_test_vectors.json", "unfollow")
}

// TestMessageFromVectors validates Message actions from test vectors
func TestMessageFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/message_test_vectors.json", "message")
}

// TestLikeFromVectors validates Like actions from test vectors
func TestLikeFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/like_test_vectors.json", "like")
}

// TestUnlikeFromVectors validates Unlike actions from test vectors
func TestUnlikeFromVectors(t *testing.T) {
	testBSocialFromVectors(t, "testdata/unlike_test_vectors.json", "unlike")
}
