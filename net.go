package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func getQueryParam(r *http.Request, name string) string {
	params := r.URL.Query()
	paramValues := params[name]
	if len(paramValues) <= 0 {
		return ""
	}

	return paramValues[0]
}

func readWebsocket(conn *websocket.Conn, cancel context.CancelFunc) {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			zlog.Info("websocket received error (closing)", zap.Error(err))
			conn.Close()
			cancel()
			return
		}
		zlog.Info("websocket received payload", zap.String("payload", string(payload)))
	}

}
func maybeSendWebsocket(conn *websocket.Conn, objType string, obj interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    objType,
		"payload": obj,
	})
	if err != nil {
		zlog.Warn("cannot marshal object", zap.String("object_type", objType), zap.Reflect("object", obj))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		zlog.Info("cannot send data", zap.String("data", string(data)))
		conn.Close()
	}
}
