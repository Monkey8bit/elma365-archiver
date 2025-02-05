package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	connectors "elma365-archiver/gateway/connectors"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

const filesBucketName = "files"

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(Response{Message: "Invalid request method", Status: http.StatusMethodNotAllowed})
		return
	}

	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to parse form", Status: http.StatusBadRequest})
		return
	}

	connector, err := connectors.CreateMinioConnector("minio:9000", accessKey, secretKey)

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to create connector", Status: http.StatusInternalServerError})
		return
	}

	files := r.MultipartForm.File["files"]
	mail := r.FormValue("mail")

	var wg sync.WaitGroup

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			json.NewEncoder(w).Encode(Response{Message: "Unable to open file", Status: http.StatusInternalServerError})
			continue
		}
		wg.Add(1)
		go serializeFile(file, fileHeader.Filename, fileHeader.Size, &wg, connector)
	}

	go func() {
		wg.Wait()
		fmt.Printf("All files sent to %s\n", mail)
	}()

	json.NewEncoder(w).Encode(Response{Message: "Files uploaded", Status: http.StatusOK})
}

func serializeFile(file multipart.File, fileName string, fileSize int64, wg *sync.WaitGroup, connector *connectors.Connector) {
	defer file.Close()
	defer wg.Done()

	log.Print(fileName)
	log.Print(fileSize)

	err, fileTag := connector.UploadFile(context.Background(), filesBucketName, file, fileName, fileSize, "application/octet-stream")

	if err != nil {
		fmt.Printf("Unable to upload file %s: %s", fileName, err)
		return
	}

	fmt.Printf("Uploaded file %s with tag %s", fileName, fileTag)
}
