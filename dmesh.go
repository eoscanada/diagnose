package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eoscanada/dmesh"
	"go.uber.org/zap"
)

func (r *Diagnose) searchPeers(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - Search Peers")

	ctx, cancel := context.WithCancel(req.Context())

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	go readWebsocket(conn, cancel)

	servicePrefix := fmt.Sprintf("%s/search", r.DmeshServiceVersion)

	zlog.Info("observing dmesh", zap.String("namespace", r.Namespace), zap.String("service_prefix", servicePrefix))
	eventChan := dmesh.Observe(ctx, r.dmeshStore, r.Namespace, servicePrefix)
	for {
		select {
		case <-ctx.Done():
			zlog.Debug("context canceled")
			return
		case peer := <-eventChan:
			maybeSendWebsocket(conn, WebsocketTypePeerEvent, peer)
		}
	}
	zlog.Info("diagnose - Search Peers - Complete")

}
