package server

import (
	"fmt"
	"net/http"

	"github.com/CGuilhem/hackathon-notifier/backend/internal/websocket"
)

func addRoutes(
	mux *http.ServeMux,
	wsManager *websocket.Manager,
	getenv func(string) string,
	logger logger,
) {

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 404 error handling
		if r.URL.Path != "/" {
			fmt.Println("404")
			logger.Error("404 error")

			return
		}
		fmt.Println("/")
	})

	// Websockets
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsManager.ServeWS(w, r)
	})
}
