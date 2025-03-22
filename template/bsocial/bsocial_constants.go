package bsocial

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

// Context defines different contexts in BSocial
type Context string

const (
	ContextTx       Context = "tx"
	ContextChannel  Context = "channel"
	ContextBapID    Context = "bapID"
	ContextProvider Context = "provider"
	ContextVideoID  Context = "videoID"
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
