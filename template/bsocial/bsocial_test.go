package bsocial

import (
	"testing"

	"github.com/bitcoinschema/go-bmap"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-templates/template/bitcom"
	"github.com/stretchr/testify/require"
)

func TestCreatePost(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Create a test post
	post := Post{
		MediaType: bitcom.MediaTypeTextMarkdown,
		Encoding:  EncodingUTF8,
		Content:   "# Hello BSV\nThis is a test post",
		Tags:      []string{"test", "bsv", "markdown"},
	}

	// Create the transaction
	tx, err := CreatePost(post, nil, nil, privKey)
	require.NoError(t, err)

	// Parse with bmap
	bmapTx, err := bmap.NewFromRawTxString(tx.String())
	require.NoError(t, err)

	// Verify MAP data
	require.NotNil(t, bmapTx.MAP)
	require.Len(t, bmapTx.MAP, 2) // One for post, one for tags

	// Find the post MAP entry
	var postMap, tagMap map[string]interface{}
	for _, m := range bmapTx.MAP {
		if m["type"] == "post" {
			postMap = m
		} else if m["CMD"] == "ADD" && m["tags"] != nil {
			tagMap = m
		}
	}

	// Verify post data
	require.NotNil(t, postMap)
	require.Equal(t, "bsocial", postMap["app"])
	require.Equal(t, "post", postMap["type"])

	// Verify tags
	require.NotNil(t, tagMap)
	tags, ok := tagMap["tags"].([]interface{})
	require.True(t, ok)
	require.Len(t, tags, len(post.Tags))
	for i, tag := range post.Tags {
		require.Equal(t, tag, tags[i])
	}

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)
	require.Equal(t, post.Content, string(bmapTx.B[0].Data))
	require.Equal(t, string(post.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(post.Encoding), bmapTx.B[0].Encoding)
}

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

func TestCreateReply(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Test txid to reply to
	testTxID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create a test reply
	reply := Post{
		MediaType: bitcom.MediaTypeTextPlain,
		Encoding:  EncodingUTF8,
		Content:   "This is a test reply",
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
	require.Equal(t, testTxID, mapData["context_tx"])

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)
	require.Equal(t, reply.Content, string(bmapTx.B[0].Data))
	require.Equal(t, string(reply.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(reply.Encoding), bmapTx.B[0].Encoding)
}

func TestCreateMessage(t *testing.T) {
	// Create a test private key
	privKey, err := ec.NewPrivateKey()
	require.NoError(t, err)

	// Create a test message
	msg := Message{
		MediaType:    bitcom.MediaTypeTextPlain,
		Encoding:     bitcom.EncodingUTF8,
		Content:      "Hello, this is a test message",
		Context:      ContextChannel,
		ContextValue: "test-channel",
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
	require.Equal(t, string(msg.Context), mapData["context_channel"])

	// Verify B data
	require.NotNil(t, bmapTx.B)
	require.Len(t, bmapTx.B, 1)
	require.Equal(t, msg.Content, string(bmapTx.B[0].Data))
	require.Equal(t, string(msg.MediaType), bmapTx.B[0].MediaType)
	require.Equal(t, string(msg.Encoding), bmapTx.B[0].Encoding)
}

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
