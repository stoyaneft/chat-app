package db

import(
	"database/sql"
	 "log"

	"gopkg.in/gorp.v1"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
    Id int64
	Username string
    Email string
    Password []byte
}

type UserChat struct {
	UserId int64
	ChatUid string
}

type Chat struct {
	Uid string
	Name string
}

type ChatMessage struct {
	UserId int64
	ChatUid string
	Message string
	CreatedAt int64
}

type Db struct { dbmap *gorp.DbMap }

func (db *Db) Init(dbPath string) {
	 sqlDb, err := sql.Open("sqlite3", dbPath)
	 checkErr(err, "sql.Open failed")

	 db.dbmap = &gorp.DbMap{Db: sqlDb, Dialect: gorp.SqliteDialect{}}
	 users := db.dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	 users.ColMap("Username").SetUnique(true)
	 db.dbmap.AddTableWithName(UserChat{}, "user_chats").SetKeys(false, "UserId", "ChatUid")
	 db.dbmap.AddTableWithName(Chat{}, "chats")

	 err = db.dbmap.CreateTablesIfNotExists()
	 checkErr(err, "Create tables failed")
}

func (db *Db) InsertUser(user User) error {
	err := db.dbmap.Insert(&user);
	return err
}

func (db *Db) SelectUser(query string, username string) (User, error) {
	var user User
	err := db.dbmap.SelectOne(&user, query, username)
	return user, err
}

func (db *Db) InsertUserChat(userChat UserChat) error {
	err := db.dbmap.Insert(&userChat);
	return err;
}

func (db *Db) SelectUsersByChat(chatUid string) ([]string, error) {
	var users []User
	_, err := db.dbmap.Select(&users, "select users.Username from user_chats join users where users.Id=user_chats.UserId and chatUid=?", chatUid)
	userIds := make([]string, len(users))
	for idx, user := range(users) {
		userIds[idx] = user.Username
	}
	return userIds, err
}

func (db *Db) SelectUserChats(username string) ([]string, error) {
	var userChats []UserChat
	_, err := db.dbmap.Select(&userChats, "select ChatUid from user_chats join users where users.Id=user_chats.UserId and users.Username=?", username)
	chatUids := make([]string, len(userChats))
	for idx, userChat := range(userChats) {
		chatUids[idx] = userChat.ChatUid
	}
	return chatUids, err
}

func (db *Db) InsertMessage(msg ChatMessage) error {
	err := db.Insert(&msg);
	return err
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
