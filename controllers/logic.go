package controllers

import (
	"net"
	"net/http"
	"encoding/json"
	// "github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type ClientManager struct{
	clients map[*Client]bool
	broadcast chan []byte
	register  chan *Client
	unregister chan *Client
}

type Client struct {
	id string
	socket *websocket.Conn
	send chan []byte
}

type Message struct{
	Sender string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"` 
	Content   string `json:"content,omitempty"`   
	ServerIP  string `json:"serverIp,omitempty"`  
	SenderIP  string `json:"senderIp,omitempty"`
}

// create a client manager
var Manager = ClientManager{
	broadcast: make(chan []byte),
	register: make(chan *Client),
	unregister: make(chan *Client),
	clients: make(map[*Client]bool),
}

func (manager *ClientManager) Start(){
	for {
		select{
		// if there is a new connection access, pass the connection to conn through the channel
		case conn := <-manager.register:
			// Set the client connection to true
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected. ", ServerIP: LocalIp(), SenderIP: conn.socket.RemoteAddr().String()})
			// Call the client's send metho and sed message
			manager.Send(jsonMessage, conn)
		// if the connection is disconnected
		case conn := <-manager.unregister:
			// Determine the state of the connection, it is true, turn off Send and delete the value of connecting client
			if _,ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected. ", ServerIP: LocalIp(), SenderIP: conn.socket.RemoteAddr().String()})
				manager.Send(jsonMessage, conn)	
			}
	
		// broadcast
		case message := <-manager.broadcast:
			// Traverse the clients that has been connected, send message to them
			for conn := range manager.clients{
				select {
					case conn.send <- message:
					default:
						close(conn.send)
						delete(manager.clients, conn)
				}
		
			}
		}
	}
}

func (manager *ClientManager) Send(message []byte, ignore *Client){
	for conn := range manager.clients{
		// Send messages no to the shielded connection
		if conn != ignore{
			conn.send <- message
		}
	}
}

func (c *Client) Read() {
	defer func() {
		Manager.unregister <- c
		_ = c.socket.Close()
		
	}()

	for {
		// read message 
		_, message, err := c.socket.ReadMessage()
		// if there is an error, cancel this connection and close it
		if err != nil {
			_= c.socket.Close()
			break
		}
		// if there isno error, put the information in broadcast
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message), ServerIP: LocalIp(), SenderIP: c.socket.RemoteAddr().String()})
		Manager.broadcast <- jsonMessage

	}

}

func (c *Client) Write() {
	defer func() {
		_ = c.socket.Close()
	}()

	for {
		select {
			// Read the message from send
		case message, ok := <-c.send:
			// if there is no message
			if !ok {
				_ = c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return 
			}
			// write it if there is new message an send it to the web
			_ = c.socket.WriteMessage(websocket.TextMessage, message)
		}
	
	}
}


var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024 * 1024 * 1024,
	WriteBufferSize: 1024 * 1024 * 1024,
	// solving cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(res http.ResponseWriter, req *http.Request){
	// Upgrade the HTTP protocol to the websocket protocol
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil{
		http.NotFound(res, req)
		return
	}

	// Every connection will open a new client, client.id generate through UUID to ensure each time it is different
	client := &Client{id: uuid.Must(uuid.NewV4(), nil).String(), socket: conn, send: make(chan []byte)}
	// Register a new link
	Manager.register <- client

	// Start the message to collect the news from the web 
	go client.Read()
	// Start the corporation to return the message to the web
	go client.Write()

}

func HealthHandler(res http.ResponseWriter,_ *http.Request){
	_, _ = res.Write([]byte("ok"))
}

func LocalIp() string {
	address, _ := net.InterfaceAddrs()
	var ip = "localhost"
	for _, address := range address {
		if ipAddress, ok := address.(*net.IPNet); ok && !ipAddress.IP.IsLoopback() {
			if ipAddress.IP.To4() != nil{
				ip = ipAddress.IP.String()
			}
		}
	}
	return ip
}