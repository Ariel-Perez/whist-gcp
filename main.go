// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_flex_websockets_app]

// Sample websockets demonstrates an App Engine Flexible app.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// Client type
type Client struct {
	ID int
	Name string
}

// Room type
type Room struct {
	Name string
	Connections map[string]*websocket.Conn
}

// Broadcast sends a byte message to all the clients in a room.
func (room Room) Broadcast(message []byte) {
	for _, connection := range room.Connections {
		if err := connection.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("conn.WriteMessage: %v", err)
		}
	}
}

// BroadcastString sends a string message to all the clients in a room.
func (room Room) BroadcastString(message string) {
	room.Broadcast([]byte(message))
}

func main() {
	room := Room{"Default", make(map[string]*websocket.Conn)}

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/ws", room.socketHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// upgrader holds the websocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// socketHandler echos websocket messages back to the client.
func (room Room) socketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	if err != nil {
		log.Printf("upgrader.Upgrade: %v", err)
		return
	}

	name := string(r.FormValue("name"))
	room.BroadcastString(name + " has joined the room.")
	defer room.BroadcastString(name + " has left the room.")

	room.Connections[name] = conn
	defer delete(room.Connections, name)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			return
		}

		if messageType == websocket.TextMessage {
			room.BroadcastString(name + " " + string(p))
		}
	}
}

// healthCheckHandler is used by App Engine Flex to check instance health.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

// [END gae_flex_websockets_app]
