package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

// Message adalah struct messaging
type Message struct {
	PrivyID string `json:"privy_id"`
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	// start goroutine to send broadcast
	go handleMessages()

	log.Println("Server running on http://localhost:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("Error server :", err)
	}

}

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade ke websocket connection
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("error connect to websocket")
	}
	// close connection when out of function
	defer ws.Close()

	// add new client when new registration
	clients[ws] = true

	for {
		var msg Message
		// read message from websocket connection
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error membaca pesan : %v", err)
			delete(clients, ws)
			break
		}

		// send message to channel broadcast
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// get message from broadcast channel
		msg := <-broadcast

		// send message to all client that connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error mengirim pesan ke klien : %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}

}
