package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func websocketCloser(conn *websocket.Conn, cancel context.CancelFunc) {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			// connection closed?
			cancel()
			return
		}
		fmt.Println("payload: ", string(payload))
	}

}

func sendMessage(conn *websocket.Conn, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	return nil
}
