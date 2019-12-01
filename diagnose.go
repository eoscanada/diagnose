package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Diagnose interface {
	SetupRoutes(s *mux.Router)
	BlockHoles(w http.ResponseWriter, r *http.Request)
	DBHoles(w http.ResponseWriter, r *http.Request)
	SearchHoles(w http.ResponseWriter, r *http.Request)
	GetBlockStoreUrl() string
	GetSearchIndexesStoreUrl() string
	GetKvdbConnectionInfo() string
	SetUpgrader(upgrader *websocket.Upgrader)
}
