package model

type User interface {
	CheckPassword([]byte) bool
	SendMessage(ChatMessage) error
	Info() UserInfo
}

type UserInfo struct {
	// TODO: uid
	ID       int64
	Username string
	Email    string
	Password []byte
}

type UserChat struct {
	// TODO: uid
	UserID  int64
	ChatUID string
}
