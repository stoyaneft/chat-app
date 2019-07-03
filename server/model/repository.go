package model

type ChatRepository interface {
	CreateUser(UserInfo) error
	GetUserByName(string) (UserInfo, error)

	GetUsersByChat(string) (UserInfo, error)
	CreateUserChat(UserChat) error

	CreateChat(UserChat) error
}
