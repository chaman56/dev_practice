package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting the server...")

	api := &api{addr: ":8080"}
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    api.addr,
		Handler: mux,
	}

	mux.HandleFunc("GET /users", api.getUsers)
	mux.HandleFunc("POST /users", api.createUser)

	server.ListenAndServe()
}
