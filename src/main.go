package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stoyaneft/chat-app/src/db"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

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
	clients[ws] = true
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		log.Printf("msg.Password:%v", msg.Password);
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		if (msg.Type != "authentication") {
			// Send it out to every client that is currently connected
			for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		} else {
			user := db.User{Username: msg.Username, Email: msg.Email, Password: msg.Password}
			// user := db.User{Username: "tosho_sexa",Email: "chocho@gmail.com",Password: "sex_bog99"}
			log.Println(user)
			err := dbx.Insert(user)
			log.Println(err)
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	dbx.Init()

	// Start listening for incoming chat messages
	go handleMessages()

	log.Println("http server started on :8001")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	// db := &db.Db{}
	// db.Insert(User{})

}
