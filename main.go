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
	Messages []db.ChatMessage
}

var chats = make(map[string]Chat) // {chatUid: Chat}
var users = make(map[string]*websocket.Conn);

var upgrader = websocket.Upgrader{}
var dbx = &db.Db{}

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
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		checkErr(err, "Could not upgrade to websockets")
	}
	defer ws.Close()
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		switch msg.Type {
		case "registration":
			handleRegistration(msg, ws)
		case "login":
			handleLogin(msg, ws)
		case "createChat":
			handleCreateChat(msg, ws)
		case "chatSelection":
			handleChatSelection(msg, ws)
		case "addUser":
			handleAddUser(msg, ws)
		case "sendMessage":
			handleSendMessage(msg)
		}
	}
}

func handleMessages(chatUid string) {
	chat := chats[chatUid]
	for msg := range chat.BroadcastQueue {
		for user := range chat.Users {
			err := user.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				user.Close()
				delete(chat.Users, user)
			}
		}
	}
}

func handleRegistration(msg Message, ws *websocket.Conn) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(msg.Password), bcrypt.DefaultCost)
	checkErr(err, "EncryptionError")
	user := db.User{
		Username: msg.Username,
		Email: msg.Email,
		Password: hashedPassword,
	}
	err = dbx.InsertUser(user)
	if (err != nil) {
		ws.WriteJSON(Message{Type: "error", Message: "User already exists"})
	} else {
		ws.WriteJSON(Message{Type: "registrationSuccessful"})
	}
}

func handleLogin(msg Message, ws *websocket.Conn) {
	log.Println("logging ", msg)
	user, err := dbx.SelectUser(msg.Username)
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
				messages, _ := dbx.SelectMessagesByChat(chatUid)
				if _, ok := chats[chatUid]; !ok {
					chats[chatUid] = Chat{
						Users: map[*websocket.Conn]bool{ws: true},
						BroadcastQueue: make(chan Message),
					}
					go handleMessages(chatUid)
				} else {
					chats[chatUid].Users[ws] = true
					log.Println(len(chats[chatUid].Users))
				}
				chatInfo := ChatInfo{Uid: chatUid, Participants: participants, Messages: messages}
				chatInfos[idx] = chatInfo
			}
			ws.WriteJSON(Message{Type: "loginSuccessful", Chats: chatInfos })
		}
	}
}

func handleCreateChat(msg Message, ws *websocket.Conn) {
	username := msg.Username
	user, err := dbx.SelectUser(username)
	chatUid := uuid.NewV4().String()
	userChat := db.UserChat{
		UserId: user.Id,
		ChatUid: chatUid,
	}
	err = dbx.InsertUserChat(userChat)
	if (err != nil) {
		log.Println(err)
		ws.WriteJSON(Message{Type: "error", Message: "Could not create chat"})
	} else {
		log.Println("chatCreationSuccessful")
		ws.WriteJSON(Message{Type: "chatCreationSuccessful", ChatUid: chatUid})
	}
	chats[chatUid] = Chat{
		Users: map[*websocket.Conn]bool{ws: true},
		BroadcastQueue: make(chan Message),
	}
	go handleMessages(chatUid);
}

func handleChatSelection(msg Message, ws *websocket.Conn) {
	usersInChat, err := dbx.SelectUsersByChat(msg.ChatUid)
	checkErr(err, "Chat selection error")
	log.Println("chat Participatnts", usersInChat)
	ws.WriteJSON(Message{Type: "chatSelectionSuccessful", Participants: usersInChat, ChatUid: msg.ChatUid})
}

func handleAddUser(msg Message, ws *websocket.Conn)  {
	user, err := dbx.SelectUser(msg.Username)
	if (err != nil) {
		log.Println(err)
		ws.WriteJSON(Message{Type: "error", Message: "User does not exist"})
		return
	}
	userChat := db.UserChat{
		UserId: user.Id,
		ChatUid: msg.ChatUid,
	}
	err = dbx.InsertUserChat(userChat)
	checkErr(err, "Insertion user chat error");
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
}

func handleSendMessage(msg Message)  {
	log.Println("Received message", msg)
	chats[msg.ChatUid].BroadcastQueue <- msg
	message := db.ChatMessage{
		Username: msg.Username,
		ChatUid: msg.ChatUid,
		Message: msg.Message,
	}
	err := dbx.InsertMessage(message)
	if (err != nil) {
		log.Println("Could not insert message: ", err)
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

func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/ws", handleConnections)
	dbx.Init("db/chat.db")

	log.Println("http server started on :8001")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
