package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"final_project/pkg/api"
	"final_project/pkg/db"

	"github.com/joho/godotenv"
)

const (
	file        = "scheduler.db"
	defaultPort = "7540"
	Password    = "12345"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbFile := strings.TrimSpace(os.Getenv("TODO_DBFILE"))
	if dbFile == "" {
		dbFile = file
		log.Printf("Variable TODO_DBFILE is not installed, default DBFILE is used: %s", dbFile)
	} else {
		log.Printf("Used DBFILE from TODO_DBFILE: %s", dbFile)
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("DB initialisation error: %v", err)
	}
	defer db.DB.Close()

	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		password = Password
		log.Printf("Variable TODO_PASSWORD is not installed, default password is used: %s", password)
	}

	api.Init(password)

	port := strings.TrimSpace(os.Getenv("TODO_PORT"))
	if port == "" {
		port = defaultPort
		log.Printf("Variable TODO_PORT is not installed, default port is used: %s", port)
	} else {
		log.Printf("Used port from TODO_PORT: %s", port)
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Port must be an integer: %s", port)
	}

	addr := ":" + port
	http.Handle("/", http.FileServer(http.Dir("web")))
	log.Printf("Start server on :%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server start error: %v", err)
	}

}
