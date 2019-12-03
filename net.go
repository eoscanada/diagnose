package main

import (
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func websocketRead(conn *websocket.Conn) {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			zlog.Info("websocket received error (closing)", zap.Error(err))
			return
		}
		zlog.Info("websocket received payload", zap.String("payload", string(payload)))
	}

}

func websocketCloser(conn *websocket.Conn, cancel context.CancelFunc) {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			zlog.Info("websocket received error (closing)", zap.Error(err))
			cancel()
			return
		}
		zlog.Info("websocket received payload", zap.String("payload", string(payload)))
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
