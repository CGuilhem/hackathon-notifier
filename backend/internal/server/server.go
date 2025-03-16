package server

import (
	"net/http"

	"github.com/CGuilhem/hackathon-notifier/backend/internal/websocket"
)

type logger interface {
	Error(msg string)
	Info(msg string)
	Warn(msg string)
}

func NewServer(
	wsManager *websocket.Manager,
	getenv func(string) string,
	logger logger,
) http.Handler {

	mux := http.NewServeMux()

	addRoutes(
		mux,
		wsManager,
		getenv,
		logger,
	)

	var handler http.Handler = mux

	return handler
}
