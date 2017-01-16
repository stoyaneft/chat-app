package db

import(
	"database/sql"
	 "log"

	"gopkg.in/gorp.v1"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
    Id      int64
	Username string
    Email string
    Password   string
}

type UserChats struct {
	UserId int64
	ChatId int64
}

type Chat struct {
	Id int64
	Name string
}

type ChatMessage struct {
	UserId int64
	ChatId int64
	Message string
	CreatedAt int64
}

type Db struct { dbmap *gorp.DbMap }

func (db *Db) Init() {
	 sqlDb, err := sql.Open("sqlite3", "db/users.db")
	 checkErr(err, "sql.Open failed")

	 db.dbmap = &gorp.DbMap{Db: sqlDb, Dialect: gorp.SqliteDialect{}}
	 db.dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	 err = db.dbmap.CreateTablesIfNotExists()
	 checkErr(err, "Create tables failed")
}

func (db *Db) Insert(user User) error {
	err := db.dbmap.Insert(&user);
	return err
}

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}
