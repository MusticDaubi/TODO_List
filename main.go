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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbFile := strings.TrimSpace(os.Getenv("TODO_DBFILE"))
	if dbFile == "" {
		dbFile = "scheduler.db"
		log.Printf("Variable TODO_DBFILE is not installed, default DBFILE is used: %s", dbFile)
	} else {
		log.Printf("Used DBFILE from TODO_DBFILE: %s", dbFile)
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("DB initialisation error: %v", err)
	}
	defer db.DB.Close()

	api.Init()

	port := strings.TrimSpace(os.Getenv("TODO_PORT"))
	if port == "" {
		port = "7540"
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
