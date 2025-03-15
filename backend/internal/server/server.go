package server

import (
	"net/http"
)

type logger interface {
	Error(msg string)
	Info(msg string)
	Warn(msg string)
}

func NewServer(
	getenv func(string) string,
	logger logger,
) http.Handler {

	mux := http.NewServeMux()

	addRoutes(
		mux,
		getenv,
		logger,
	)

	var handler http.Handler = mux

	return handler
}
