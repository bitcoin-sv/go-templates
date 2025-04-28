package bsocial

// ActionType defines different action types in BSocial
type ActionType string

const (
	// Action types
	TypePostReply ActionType = "post" // Used for both posts and replies
	TypeLike      ActionType = "like"
	TypeUnlike    ActionType = "unlike"
	TypeFollow    ActionType = "follow"
	TypeUnfollow  ActionType = "unfollow"
	TypeMessage   ActionType = "message"
)

// ActionContext defines different contexts in BSocial
type ActionContext string

const (
	// used for replies
	ContextTx ActionContext = "tx"
	// used for posts, messages, etc.
	ContextChannel  ActionContext = "channel"
	ContextBapID    ActionContext = "bapID"
	ContextProvider ActionContext = "provider"
	ContextVideoID  ActionContext = "videoID"
	ContextGeohash  ActionContext = "geohash"
	ContextBtcTx    ActionContext = "btcTx"
	ContextEthTx    ActionContext = "ethTx"
)

// IsEmpty checks if a BSocial object is empty (has no content)
func (bs *BSocial) IsEmpty() bool {
	return bs.Post == nil &&
		bs.Reply == nil &&
		bs.Like == nil &&
		bs.Unlike == nil &&
		bs.Follow == nil &&
		bs.Unfollow == nil &&
		bs.Message == nil &&
		bs.AIP == nil &&
		len(bs.Attachments) == 0 &&
		len(bs.Tags) == 0
}
