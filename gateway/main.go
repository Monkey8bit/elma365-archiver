package main

import (
	"elma365-archiver/gateway/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/files", handlers.UploadFileHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
