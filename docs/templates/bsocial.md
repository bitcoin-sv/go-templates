# BSocial Template

## Overview

The BSocial template provides a standardized way to create social media transactions on the Bitcoin SV blockchain using composable data protocols like "B", Magic Attribute Protocol (MAP), Author Identity Protocol (AIP), and Bitcoin Attestation Protocol (BAP). It defines common actions such as posts, replies, likes, follows, and messages.

BSocial implements the social graph protocols defined by [BitcoinSchema.org](https://bitcoinschema.org/), a community-driven initiative creating extensible schemas that enable developers to build interoperable data applications on Bitcoin.

## Features

- Create posts with Markdown or plain text content
- Reply to existing posts
- Like and unlike posts
- Follow and unfollow users
- Send messages in channels
- Add tags to posts
- Sign transactions with AIP (Author Identity Protocol) for authentication
- Full compatibility with BitcoinSchema.org protocols

## BitcoinSchema.org Integration

This template implements the following BitcoinSchema.org protocols:

- **MAP** (Magic Attribute Protocol) - For structured key-value data storage
- **B** - For binary data and content attachments
- **AIP** (Author Identity Protocol) - Signing data with an identity
- **BAP** (Bitcoin Attestation Protocol) - Registering / managing on-chain identities

The BSocial template is part of the broader BitcoinSchema ecosystem, which includes libraries like:
- [go-map](https://github.com/BitcoinSchema/go-map) - Go library for working with Magic Attribute Protocol
- [go-aip](https://github.com/BitcoinSchema/go-aip) - Go library for working with Author Identity Protocol
- [go-bap](https://github.com/BitcoinSchema/go-bap) - Go library for working with Bitcoin Attestation Protocol
- [go-b](https://github.com/BitcoinSchema/go-b) - Go library for working with B data protocol
- [go-bmap](https://github.com/BitcoinSchema/go-bmap) - Go library for working with Bitcoin data protocols

## Installation

```bash
go get github.com/bitcoin-sv/go-templates/template/bsocial
```

## Usage Examples

### Creating a Post

```go
import (
    "github.com/bitcoin-sv/go-templates/template/bsocial"
    "github.com/bitcoin-sv/go-templates/template/bitcom"
    ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

// Create a private key (or use an existing one)
privKey, err := ec.NewPrivateKey()
if err != nil {
    // Handle error
}

// Create a post
post := bsocial.Post{
    B: bitcom.B{
        MediaType: bitcom.MediaTypeTextMarkdown,
        Encoding:  bitcom.EncodingUTF8,
        Data:      []byte("# Hello BSV\nThis is my first post using BSocial!"),
    },
    Action: bsocial.Action{
        Type: bsocial.TypePostReply,
    },
}

// Add optional tags
tags := []string{"bitcoin", "bsv", "onchain"}

// Create the transaction
tx, err := bsocial.CreatePost(post, nil, tags, privKey)
if err != nil {
    // Handle error
}

// tx can now be processed, signed, broadcast, etc.
```

### Liking a Post

```go
// Like an existing post by its transaction ID
txID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
tx, err := bsocial.CreateLike(txID, nil, nil, privKey)
if err != nil {
    // Handle error
}

// tx can now be processed, signed, broadcast, etc.
```

### Replying to a Post

```go
// Reply to an existing post
replyToTxID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

reply := bsocial.Reply{
    B: bitcom.B{
        MediaType: bitcom.MediaTypeTextPlain,
        Encoding:  bitcom.EncodingUTF8,
        Data:      []byte("Great post! I completely agree."),
    },
    Action: bsocial.Action{
        Type:         bsocial.TypePostReply,
        Context:      bsocial.ContextTx,
        ContextValue: replyToTxID,
    },
}

tx, err := bsocial.CreateReply(reply, replyToTxID, nil, nil, privKey)
if err != nil {
    // Handle error
}

// tx can now be processed, signed, broadcast, etc.
```

### Following a User

```go
// Follow a user by their BAP ID
bapID := "user-bap-identifier"
tx, err := bsocial.CreateFollow(bapID, nil, nil, privKey)
if err != nil {
    // Handle error
}

// tx can now be processed, signed, broadcast, etc.
```

### Sending a Message to a Channel

```go
// Send a message to a channel
message := bsocial.Message{
    B: bitcom.B{
        MediaType: bitcom.MediaTypeTextPlain,
        Encoding:  bitcom.EncodingUTF8,
        Data:      []byte("Hello everyone in this channel!"),
    },
    Action: bsocial.Action{
        Type:         bsocial.TypeMessage,
        Context:      bsocial.ContextChannel,
        ContextValue: "channel-name",
    },
}

tx, err := bsocial.CreateMessage(message, nil, nil, privKey)
if err != nil {
    // Handle error
}

// tx can now be processed, signed, broadcast, etc.
```

### Decoding BSocial Transactions

```go
// Assuming 'tx' is a transaction obtained from the network
bsocialData := bsocial.DecodeTransaction(tx)

// Check what type of action it contains
if bsocialData.Post != nil {
    // This is a post
    postContent := string(bsocialData.Post.B.Data)
    // Process the post...
} else if bsocialData.Reply != nil {
    // This is a reply
    replyContent := string(bsocialData.Reply.B.Data)
    originalPostTxID := bsocialData.Reply.ContextValue
    // Process the reply...
} else if bsocialData.Like != nil {
    // This is a like
    likedPostTxID := bsocialData.Like.ContextValue
    // Process the like...
}
```

## Data Structures

### BSocial Types

```go
// BSocialType defines different action types in BSocial
type BSocialType string

const (
    // Action types
    TypePostReply BSocialType = "post" // Used for both posts and replies
    TypeLike      BSocialType = "like"
    TypeUnlike    BSocialType = "unlike"
    TypeFollow    BSocialType = "follow"
    TypeUnfollow  BSocialType = "unfollow"
    TypeMessage   BSocialType = "message"
)
```

### Common Context Types

A context can be any string that is meaningful to the action. Here are some common contexts:
```go
// Context defines different contexts in BSocial
type Context string

const (
    ContextTx       Context = "tx"       // Transaction context
    ContextChannel  Context = "channel"  // Channel context
    ContextBapID    Context = "bapID"    // Bitcoin Attestation Protocol ID
    ContextProvider Context = "provider" // Provider context
    ContextVideoID  Context = "videoID"  // Video ID context
    ContextGeohash  Context = "geohash"  // Geohash context
)
```

## Additional Resources

- [BitcoinSchema.org](https://bitcoinschema.org/) - Standards for on-chain data formats and protocols
- [BitcoinSchema GitHub](https://github.com/BitcoinSchema) - Open source libraries and tools
- [AIP](https://github.com/b-open-io/aip/) - Author Identity Protocol
- [B Protocol](https://github.com/bitcoinschema/go-b) - For understanding the B data protocol used in BSocial
- [MAP Protocol](https://github.com/rohenaz/MAP) - Magic Attribute Protocol documentation

## Note

The BSocial template is designed to be compliant with BitcoinSchema.org standards. By following these standards, you ensure compatibility with other applications in the Bitcoin SV ecosystem that also adhere to BitcoinSchema.org protocols. 