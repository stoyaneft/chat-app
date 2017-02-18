package db

import (
	"fmt"
	"os"
	"testing"
	//"gopkg.in/gorp.v1"
)

var dbx = &Db{}

func TestMain(m *testing.M) {
	dbx.Init("test.db")
	retCode := m.Run()

	dbx.dbmap.Db.Close()
	err := os.Remove("test.db")
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(retCode)

}

func TestDb_Init(t *testing.T) {
	if _, err := os.Stat("test.db"); os.IsNotExist(err) {
		t.Errorf("Error: db is not created")
	}
}

//
func TestDb_InsertUser(t *testing.T) {
	if err := dbx.InsertUser(User{Username: "username", Email: "email", Password: []byte("pass")}); err != nil {
		t.Errorf("Error: could not insert user ", err)
	}

	var user User
	if err := dbx.dbmap.SelectOne(&user, "select * from users where username=?", "username"); err != nil {
		t.Errorf("Error: user does not exist")
	}
	if user.Username != "username" || user.Email != "email" || string(user.Password) != "pass" {
		t.Errorf("Error: user does not match")
	}
}

func TestDb_SelectUser(t *testing.T) {
	var user User
	var err error
	if user, err = dbx.SelectUser("username"); err != nil {
		t.Errorf("Error: user does not exist")
	}
	if user.Username != "username" || user.Email != "email" || string(user.Password) != "pass" {
		t.Errorf("Error: user does not match")
	}
}

func TestDb_InsertMessage(t *testing.T) {
	if err := dbx.InsertMessage(ChatMessage{Username: "username", ChatUid: "uid", Message: "message"}); err != nil {
		t.Errorf("Error: could not insert message")
	}

	var msg ChatMessage
	if err := dbx.dbmap.SelectOne(&msg, "select * from messages where username=?", "username"); err != nil {
		t.Errorf("Error: message does not exist")
	}
	if msg.Username != "username" || msg.ChatUid != "uid" || msg.Message != "message" {
		t.Errorf("Error: message does not match")
	}
}

func TestDb_SelectMessagesByChat(t *testing.T) {
	var msgs []ChatMessage
	var err error
	if msgs, err = dbx.SelectMessagesByChat("uid"); err != nil {
		t.Errorf("Error: messages does not exist")
	}
	msg := msgs[0]
	if msg.Username != "username" || msg.ChatUid != "uid" || msg.Message != "message" {
		t.Errorf("Error: message does not match")
	}
}

func TestDb_InsertUserChat(t *testing.T) {
	if err := dbx.InsertUserChat(UserChat{UserId: 1, ChatUid: "uid"}); err != nil {
		t.Errorf("Error: could not insert user")
	}

	var chat UserChat
	if err := dbx.dbmap.SelectOne(&chat, "select * from user_chats where ChatUid=?", "uid"); err != nil {
		t.Errorf("Error: chat does not exist")
	}
	if chat.UserId != 1 || chat.ChatUid != "uid" {
		t.Errorf("Error: chat does not match")
	}
}

func TestDb_SelectUserChats(t *testing.T) {
	var chatUids []string
	var err error
	if chatUids, err = dbx.SelectUserChats("username"); err != nil {
		t.Errorf("Error: chats does not exist")
	}
	if chatUids[0] != "uid" {
		t.Errorf("Error: chat does not match")
	}
}

func TestDb_SelectUsersByChat(t *testing.T) {
	var users []string
	var err error
	if users, err = dbx.SelectUsersByChat("uid"); err != nil {
		t.Errorf("Error: chats does not exist")
	}
	if users[0] != "username" {
		t.Errorf("Error: chat does not match")
	}
}
