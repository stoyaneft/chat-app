package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)



type Db struct{ dbmap *gorp.DbMap }

// Creates db in the file pointed by dbPath and initializes tables
func (db *Db) Init(dbPath string) {
	sqlDb, err := sql.Open("sqlite3", dbPath)
	checkErr(err, "sql.Open failed")

	db.dbmap = &gorp.DbMap{Db: sqlDb, Dialect: gorp.SqliteDialect{}}
	users := db.dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	users.ColMap("Username").SetUnique(true)
	db.dbmap.AddTableWithName(UserChat{}, "user_chats").SetKeys(false, "UserId", "ChatUid")
	db.dbmap.AddTableWithName(Chat{}, "chats")
	db.dbmap.AddTableWithName(ChatMessage{}, "messages").SetKeys(false)

	err = db.dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
}

// Inserts user in table users
func (db *Db) InsertUser(user User) error {
	err := db.dbmap.Insert(&user)
	return err
}

// Returns user by given username
func (db *Db) SelectUser(username string) (User, error) {
	var user User
	err := db.dbmap.SelectOne(&user, "select * from users where username=?", username)
	return user, err
}

// Inserts user chat
func (db *Db) InsertUserChat(userChat UserChat) error {
	err := db.dbmap.Insert(&userChat)
	return err
}

// Returns slice of users which are in chat with uid `chatUid`
func (db *Db) SelectUsersByChat(chatUid string) ([]string, error) {
	var users []User
	_, err := db.dbmap.Select(&users, "select users.Username from user_chats join users where users.Id=user_chats.UserId and chatUid=?", chatUid)
	userIds := make([]string, len(users))
	for idx, user := range users {
		userIds[idx] = user.Username
	}
	return userIds, err
}

// Returns slice of chats for the given users with `username`
func (db *Db) SelectUserChats(username string) ([]string, error) {
	var userChats []UserChat
	_, err := db.dbmap.Select(&userChats, "select ChatUid from user_chats join users where users.Id=user_chats.UserId and users.Username=?", username)
	chatUids := make([]string, len(userChats))
	for idx, userChat := range userChats {
		chatUids[idx] = userChat.ChatUid
	}
	return chatUids, err
}

// Inserts chat message in table messages
func (db *Db) InsertMessage(msg ChatMessage) error {
	msg.CreatedAt = time.Now().UnixNano()
	err := db.dbmap.Insert(&msg)
	return err
}

// Returns all messages for given chat
func (db *Db) SelectMessagesByChat(chatUid string) ([]ChatMessage, error) {
	var messages []ChatMessage
	_, err := db.dbmap.Select(&messages, "select * from messages where ChatUid=? order by CreatedAt", chatUid)
	return messages, err
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
