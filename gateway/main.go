package main

import (
	"elma365-archiver/gateway/handlers"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	log.Println(filepath.Join("/app", ".env"))
	if err := godotenv.Load(filepath.Join("/app", ".env")); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	http.HandleFunc("/files", handlers.UploadFileHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
