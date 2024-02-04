package main

import (
	"fmt"	
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/Qwerci/Chat_app_v2/controllers"
)


// Upgrade web socket connection

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main(){
	fmt.Println("Starting application...")
	// Open a goroutine exection start program
	go controllers.Manager.Start()
	// Register the default route to /ws, and use the wsHandler method
	http.HandleFunc("/ws", controllers.WsHandler)
	http.HandleFunc("/health", controllers.HealthHandler)
	// Surveying the local 8011 port
	fmt.Println("chat server start.....")
	//Note that this must be 0.0.0.0 to deploy in the server to use
	_ = http.ListenAndServe("0.0.0.0:8448", nil)
}
