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

type Db struct { dbmap *gorp.DbMap }

func (db *Db) Init() {
	 sqlDb, err := sql.Open("sqlite3", "db/users.db")
	 checkErr(err, "sql.Open failed")

	 db.dbmap = &gorp.DbMap{Db: sqlDb, Dialect: gorp.SqliteDialect{}}
	 db.dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	 err = db.dbmap.CreateTablesIfNotExists()
	 checkErr(err, "Create tables failed")

	//  return dbmap
}

func (db *Db) Insert(user User) error {
	// log.Println(*user)
	err := db.dbmap.Insert(&user);
	return err
}

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

// func main() {
// 	// dbmap := Init()
// 	db := &Db{}
// 	db.Init()
// 	defer db.dbmap.Db.Close()
// 	chocho := &User{Username: "goocho_sexa",Email: "chocho@gmail.com",Password: "sex_bog99"}
// 	err := db.Insert(chocho);
// 	checkErr(err, "Insertion failed")
// }
