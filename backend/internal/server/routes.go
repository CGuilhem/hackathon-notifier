package server

import (
	"fmt"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
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
}
