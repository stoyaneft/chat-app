package main

import (
	"log"
	"net/http"
	"io/ioutil"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/websocket"
	"github.com/stoyaneft/chat-app/db"
)

type Chat struct {
	Uid string
	Users map[*websocket.Conn]bool
	BroadcastQueue chan Message
}

type ChatInfo struct {
	Uid string
	Participants []string
	Messages []string
}

var chats = make(map[string]Chat) // {chatUid: Chat}
var users = make(map[string]*websocket.Conn);

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
	ChatUid string `json:"chatUid"`
	Participants []string `json:"participants"`
	Chats []ChatInfo `json:"chats"`
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
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(msg.Password), bcrypt.DefaultCost)
			checkErr(err, "EncryptionError")
			user := db.User{
				Username: msg.Username,
				Email: msg.Email,
				Password: hashedPassword,
			}
			log.Println("inserted ", user);
			err = dbx.InsertUser(user)
			if (err != nil) {
				ws.WriteJSON(Message{Type: "error", Message: "User already exists"})
			} else {
				ws.WriteJSON(Message{Type: "registrationSuccessful"})
			}
		case "login":
			log.Println("logging ", msg)
			user, err := dbx.SelectUser("select * from users where username=?", msg.Username)
			log.Println(user)
			if (err != nil) {
				log.Println(err)
				ws.WriteJSON(Message{Type: "error", Message: "Username does not exist"})
			} else {
				err = bcrypt.CompareHashAndPassword(user.Password, []byte(msg.Password))
				if (err != nil) {
					ws.WriteJSON(Message{Type: "error", Message: "Wrong password"})
				} else {
					log.Println("loginSuccessful")
					users[msg.Username] = ws;
					userChats, _ := dbx.SelectUserChats(msg.Username)
					chatInfos := make([]ChatInfo, len(userChats))
					for idx, chatUid := range userChats {
						participants, _ := dbx.SelectUsersByChat(chatUid)
						if _, ok := chats[chatUid]; !ok {
							log.Println("creating chat")
							chats[chatUid] = Chat{
								Users: map[*websocket.Conn]bool{ws: true},
								BroadcastQueue: make(chan Message),
							}
							go handleMessages(chatUid)
						} else {
							log.Println("Adding to chat")
							chats[chatUid].Users[ws] = true
							log.Println(len(chats[chatUid].Users))
						}
						chatInfo := ChatInfo{Uid: chatUid, Participants: participants}
						chatInfos[idx] = chatInfo
					}
					ws.WriteJSON(Message{Type: "loginSuccessful", Chats: chatInfos })
				}
			}
		case "createChat":
			username := msg.Username
			user, err := dbx.SelectUser("select * from users where username=?", username)
			chatUid := uuid.NewV4().String()
			log.Println(chatUid)
			userChat := db.UserChat{
				UserId: user.Id,
				ChatUid: chatUid,
			}
			err = dbx.InsertUserChat(userChat)

			checkErr(err, "Insertion user chat error");
			log.Println(user)
			chats[chatUid] = Chat{
				Users: map[*websocket.Conn]bool{ws: true},
				BroadcastQueue: make(chan Message),
			}
			if (err != nil) {
				log.Println(err)
				ws.WriteJSON(Message{Type: "error", Message: "Could not create chat"})
			} else {
				log.Println("chatCreationSuccessful")
				ws.WriteJSON(Message{Type: "chatCreationSuccessful", ChatUid: chatUid})
			}
			go handleMessages(chatUid);
		case "chatSelection":
			 log.Println("selection msg", msg)
		 	usersInChat, err := dbx.SelectUsersByChat(msg.ChatUid)
			checkErr(err, "Chat selection error")
		 	log.Println("chat Participatnts", usersInChat)
			ws.WriteJSON(Message{Type: "chatSelectionSuccessful", Participants: usersInChat, ChatUid: msg.ChatUid})
		case "addUser":
			user, err := dbx.SelectUser("select * from users where username=?", msg.Username)
			if (err != nil) {
				log.Println(err)
				ws.WriteJSON(Message{Type: "error", Message: "User does not exist"})
				continue
			}
			userChat := db.UserChat{
				UserId: user.Id,
				ChatUid: msg.ChatUid,
			}
			err = dbx.InsertUserChat(userChat)
			checkErr(err, "Insertion user chat error");
			//ws.WriteJSON(Message{Type: "userAddedSuccessful", Username: user.Username})
			log.Println("Added user ", user);
			usersInChat, err := dbx.SelectUsersByChat(msg.ChatUid)
			checkErr(err, "Chat selection error")
			for _, chatMember := range usersInChat {
				if chatMemberWs, ok := users[chatMember]; ok {
					chatMemberWs.WriteJSON(Message{Type: "userAddedSuccessful", Username: user.Username})
				}
			}
			if addedUserWs, ok := users[msg.Username]; ok {
				addedUserWs.WriteJSON(Message{Type: "addedToChat", ChatUid: msg.ChatUid, Participants: usersInChat})
				chats[msg.ChatUid].Users[addedUserWs] = true
			}
		case "sendMessage":
			log.Println("Received message", msg)
			chats[msg.ChatUid].BroadcastQueue <- msg
			err := dbx.InsertMessage(msg)
		}
	}
}

func handleMessages(chatUid string) {
	chat := chats[chatUid]
	log.Println(chat)
	for msg := range chat.BroadcastQueue {
		log.Println("opaaaa", msg)
		// Grab the next message from the broadcast channel
		for user := range chat.Users {
			err := user.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				user.Close()
				delete(chat.Users, user)
			}
			log.Println("Message sent")
		}
	}
}

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
