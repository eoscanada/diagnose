package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eoscanada/dmesh"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (r *Diagnose) searchPeers(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - searchPeers")

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var ctx, cancel = context.WithCancel(req.Context())

	servicePrefix := fmt.Sprintf("%s/search", r.DmeshServiceVersion)

	go websocketCloser(conn, cancel)

	zlog.Info("observing dmesh", zap.String("namespace", r.Namespace), zap.String("service_prefix", servicePrefix))
	eventChan := dmesh.Observe(ctx, r.dmeshStore, r.Namespace, servicePrefix)
	for {
		select {
		case <-ctx.Done():
			break
		case peer := <-eventChan:
			zlog.Debug("dmesh received peer event", zap.Reflect("peer_event", peer))
			data, err := json.Marshal(peer)
			if err != nil {
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}
