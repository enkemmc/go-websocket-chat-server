package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type chatter struct {
	Username string
	IP       string
	Conn     *websocket.Conn
}

// dont broadcast connection info
type safechatter struct {
	Username string `json:"name"`
	IP       string `json:"ip"`
}

// OutgoingMessage Represents outgoing JSON to connected chat clients
type OutgoingMessage struct {
	Username          string        `json:"name"`
	IP                string        `json:"ip"`
	Message           string        `json:"msg"`
	ConnectedChatters []safechatter `json:"connectedChatters"`
}

// IncomingMessage represents form of incoming JSON from chat server for message broadcasts and new connections
type IncomingMessage struct {
	Username      string `json:"name"`
	Justconnected bool   `json:"justConnected"`
	Message       string `json:"message"`
}

var chatters = make(map[*websocket.Conn]*chatter)
var broadcast = make(chan *OutgoingMessage)

func main() {
	// routes
	http.HandleFunc("/", serveWebsite)
	http.HandleFunc("/v1/ws", handleSocket)

	go startBroadcaster()
	startServer()
}

func serveWebsite(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	locadd := conn.UnderlyingConn().LocalAddr().String()
	ipaddy, _, err := net.SplitHostPort(locadd)
	if err != nil {
		fmt.Printf("Unreadable IP address: %s\n", ipaddy)
		ipaddy = "0"
	}

	newchatter := chatter{
		Username: "",
		IP:       ipaddy,
		Conn:     conn,
	}
	chatters[conn] = &newchatter

	go socketReader(conn)
}

func socketReader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(68, err)
			broadcast <- &OutgoingMessage{
				Username: chatters[conn].Username,
				IP:       chatters[conn].IP,
				Message:  "Seeya.",
			}
			return
		}
		var incomingmessage IncomingMessage
		err = json.Unmarshal(p, &incomingmessage)
		if err != nil {
			fmt.Println("Error Unmarshalling json")
			fmt.Println(err)
			return
		}

		// save username to chatters list
		chatters[conn].Username = incomingmessage.Username

		message := OutgoingMessage{
			Username:          incomingmessage.Username,
			IP:                chatters[conn].IP,
			Message:           incomingmessage.Message,
			ConnectedChatters: *getConnectedChattersAsJSON(),
		}

		if incomingmessage.Justconnected == true {
			if message.Username == "" {
				message.Message = fmt.Sprintf("%s has connected.", message.IP)
			} else {
				message.Message = fmt.Sprintf("%s has connected.", message.Username)
			}
		}

		broadcast <- &message
	}
}

func startServer() {
	port := os.Getenv("PORT")
	fmt.Printf("Listening on %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}
}

func startBroadcaster() {
	for {
		val := <-broadcast
		fmt.Printf("chatters connected: %d\n", len(chatters))
		jmsg, err := json.Marshal(val)
		fmt.Printf("msg to send is:%s\n", jmsg)
		if err != nil {
			fmt.Println("Error marshalling json")
			return
		}
		for chatter := range chatters {
			err := chatter.WriteMessage(websocket.TextMessage, jmsg)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				chatter.Close()
				delete(chatters, chatter)
			}
		}
	}
}

func getConnectedChattersAsJSON() *[]safechatter {
	var chatterArr []safechatter
	for _, value := range chatters {

		chatterArr = append(chatterArr, safechatter{
			Username: value.Username,
			IP:       value.IP,
		})
	}
	return &chatterArr
}
