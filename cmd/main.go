package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/matthewgaim/loudmouth_api/internal/db"
	"github.com/matthewgaim/loudmouth_api/internal/ws"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file: %s\n", err.Error())
	}
	db.ConnectToDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})
	mux.HandleFunc("/ws", ws.HandleWebSocket())

	log.Println("Server starting on localhost:8443")
	if err := http.ListenAndServe(":8443", mux); err != nil {
		log.Println(err.Error())
	}
}
