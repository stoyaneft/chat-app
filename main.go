package main

import (
	"log"
	"net/http"
	"io/ioutil"

	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/websocket"
	"github.com/stoyaneft/chat-app/db"
)

type Chat struct {
	Id int64
	Users map[*User]bool
	BroadcastQueue chan Message
}

type User struct {
	Username string
	Password string
	Ws *websocket.Conn
}

var chats = make(map[int64]*Chat) // {chatId: Chat}
var clients = make(map[string]*websocket.Conn) // {username: conn}
// var clients = make(map[*websocket.Conn]bool) // connected clients
// var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{}
var dbx = &db.Db{}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Message  string `json:"message"`
	Type string `json:"type"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// clients[ws] = true
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			// delete(clients, ws)
			break
		}
		switch msg.Type {
			case "registration":
				clients[msg.Username] = ws
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(msg.Password), bcrypt.DefaultCost)
				checkErr(err, "EncryptionError")
				user := db.User{
					Username: msg.Username,
					Email: msg.Email,
					Password: hashedPassword,
				}
				log.Println("inserted ", user);
				err = dbx.Insert(user)
				if (err != nil) {
					ws.WriteJSON(Message{Type: "error", Message: "User already exists"})
				} else {
					ws.WriteJSON(Message{Type: "registrationSuccessful"})
				}
			case "login":
				clients[msg.Username] = ws
				log.Println("logging ", msg)
				user, err := dbx.Select("select * from users where username=?", msg.Username)
				log.Println(user)
				if (err != nil) {
					ws.WriteJSON(Message{Type: "error", Message: "Username does not exist"})
				} else {
					err = bcrypt.CompareHashAndPassword(user.Password, []byte(msg.Password))
					if (err != nil) {
						ws.WriteJSON(Message{Type: "error", Message: "Wrong password"})
					} else {
						log.Println("loginSuccessful")
						ws.WriteJSON(Message{Type: "loginSuccessful"})
					}
				}

			// default:
			// 	for client := range clients {
			// 		err := client.Ws.WriteJSON(msg)
			// 		if err != nil {
			// 			log.Printf("error: %v", err)
			// 			client.Close()
			// 			delete(clients, client)
			// 		}
			// 	}
			}
		// Send the newly received message to the broadcast channel
	}
}

// func handleMessages() {
// 	for {
// 		// Grab the next message from the broadcast channel
// 		msg := <-broadcast
//
// 		switch msg.Type {
// 		case "registration":
// 			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(msg.Password), bcrypt.DefaultCost)
// 			checkErr(err, "EncryptionError")
// 			user := db.User{
// 				Username: msg.Username,
// 				Email: msg.Email,
// 				Password: hashedPassword,
// 			}
// 			err = dbx.Insert(user)
// 			checkErr(err, "Insertion error")
// 		case "authentication":
//
//
// 		default:
// 			for client := range clients {
// 				err := client.Ws.WriteJSON(msg)
// 				if err != nil {
// 					log.Printf("error: %v", err)
// 					client.Close()
// 					delete(clients, client)
// 				}
// 			}
// 		}
//
// 		if (msg.Type != "authentication") {
// 			// Send it out to every client that is currently connected
//
// 		} else {
// 			// validateCredentials(msg)
//
// 		}
// 	}
// }

func loadPage(filename string) ([]byte, error) {
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return body, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := loadPage("public/login.html")
	checkErr(err, "Error loading page login.html")
	w.Write(body)
}

// Validates user credentials
func validateCredentials(msg Message) bool {
 	return true
}

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	http.HandleFunc("/login/", loginHandler)
	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	dbx.Init("db/chat.db")

	// Start listening for incoming chat messages
	// go handleMessages()

	log.Println("http server started on :8001")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	// db := &db.Db{}
	// db.Insert(User{})

}
