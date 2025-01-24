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
		log.Fatalf("Error loading environment variables: %v", err.Error())
	}
	db.ConnectToDatabase()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})
	mux.HandleFunc("/ws", ws.HandleWebSocket())
	log.Println("Server starting on localhost:8000")
	if err := http.ListenAndServe("localhost:8000", mux); err != nil {
		log.Println(err.Error())
	}
}
