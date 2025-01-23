package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/matthewgaim/loudmouth_api/internal/auth"
	"github.com/matthewgaim/loudmouth_api/internal/comments"
	"github.com/matthewgaim/loudmouth_api/internal/db"
	"github.com/matthewgaim/loudmouth_api/internal/ws"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err.Error())
	}

	pg, err := db.ConnectToDatabase()
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err.Error())
	}
	redis := db.ConnectToRedis()

	defer pg.Close()
	defer redis.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})

	mux.HandleFunc("POST /signup", auth.Signup(pg))
	mux.HandleFunc("POST /signin", auth.Signin(pg))
	mux.HandleFunc("GET /get-comments", comments.GetComments(pg))
	mux.HandleFunc("POST /make-comment", comments.MakeComment(pg))
	mux.HandleFunc("/ws", ws.HandleWebSocket)
	log.Println("Server starting on localhost:8000")
	if err := http.ListenAndServe("localhost:8000", mux); err != nil {
		log.Println(err.Error())
	}
}
