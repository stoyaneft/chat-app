package chat

import (
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/stoyaneft/chat-app/server/model"
)

type chatUseCase struct {
	repo model.ChatRepository

	chats map[string]model.ChatRoom
}

func NewChatUseCase(repo model.ChatRepository) chatUseCase {
	return chatUseCase{
		repo: repo,
	}
}

func (c *chatUseCase) RegisterUser(userInfo model.UserInfo) error {
	return c.repo.CreateUser(userInfo)
}

func (c *chatUseCase) LoginUser(user model.User) (*model.UserInfo, error) {
	userInfo := user.Info()
	existingUser, err := c.repo.GetUserByName(userInfo.Username)
	if err != nil {
		return nil, fmt.Errorf("could not login user: %s", err)
	}

	// TODO: Expose interface for checking passwords.
	// This way an arbitrary security pass mechanism can be used.
	if string(userInfo.Password) != string(existingUser.Password) {
		return nil, fmt.Errorf("could not login user: wrong password")
	}

	return &existingUser, nil
}

func (c *chatUseCase) CreateChat(username string) error {
	user, err := c.repo.GetUserByName(username)
	chatUID := uuid.NewV4().String()
	userChat := model.UserChat{
		UserID:  user.ID,
		ChatUID: chatUID,
	}

	err = c.repo.CreateChat(userChat)
	if err != nil {
		return fmt.Errorf("could not create chat: %s", err)
	}

	// TODO: lock
	c.chats[chatUID] = model.ChatRoom{
		Users:        map[model.User]struct{}{},
		MessageQueue: make(chan model.ChatMessage),
	}

	go c.handleMessages(chatUID)
	// TODO: handle system messages
	return nil
}

func (c *chatUseCase) AddUser(chatUID string, user model.UserInfo) error {
	user, err := c.repo.GetUserByName(user.Username)
	if err != nil {
		return fmt.Errorf("could not add user to chat: %s", err)
	}
	userChat := model.UserChat{
		UserID:  user.ID,
		ChatUID: chatUID,
	}

	err = c.repo.CreateUserChat(userChat)
	if err != nil {
		return fmt.Errorf("coult not create user chat: %s", err)
	}

	chat, ok := c.chats[chatUID]
	if !ok {
		return fmt.Errorf("chat %s does not exist", chatUID)
	}
	chat.SystemQueue <- model.SystemMessage{
		Type:    model.SystemUserAdded,
		UserID:  user.ID,
		ChatUID: chatUID,
	}
	return nil
}

func (c *chatUseCase) handleMessages(chatUID string) {
	// TODO: lock
	chat := c.chats[chatUID]
	for msg := range chat.MessageQueue {
		for user := range chat.Users {
			err := user.SendMessage(msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(chat.Users, user)
			}
		}
	}
}
