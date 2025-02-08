package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	// "log"
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

type FileValidationMeta struct {
	FileTag  string
	Status   bool
	FileName string
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

	files := r.MultipartForm.File["files"]
	mail := r.FormValue("mail")

	if len(mail) == 0 {
		json.NewEncoder(w).Encode(Response{Message: "Mail is required", Status: http.StatusBadRequest})
		return
	}

	connector, err := connectors.CreateMinioConnector("minio:9000", accessKey, secretKey)

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to create connector", Status: http.StatusInternalServerError})
		return
	}

	var wg sync.WaitGroup
	fileStatuses := make(chan FileValidationMeta, len(files))

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			json.NewEncoder(w).Encode(Response{Message: "Unable to open file", Status: http.StatusInternalServerError})
			continue
		}
		wg.Add(1)
		go serializeFile(file, fileHeader.Filename, fileHeader.Size, &wg, connector, fileStatuses)
	}

	go func() {
		wg.Wait()
		close(fileStatuses)
	}()

	postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB")

	postgresConnector, err := connectors.CreatePostgresConnector(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb))

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to create postgres connection", Status: http.StatusInternalServerError})
		return
	}

	userId, err := postgresConnector.GetOrCreateUser(context.Background(), mail)

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to get or create user", Status: http.StatusInternalServerError})
		return
	}

	var filesForInsert []FileValidationMeta

	for file := range fileStatuses {
		if file.Status {
			filesForInsert = append(filesForInsert, file)
		}
	}

	filesIds := []int{}

	for _, file := range filesForInsert {
		fmt.Println(file)
		fileId, err := postgresConnector.InsertFile(context.Background(), connectors.FileMeta{FileName: file.FileName, MinioTag: file.FileTag, UserId: userId})

		if err != nil {
			json.NewEncoder(w).Encode(Response{Message: "Unable to insert file", Status: http.StatusInternalServerError})
			return
		}

		filesIds = append(filesIds, fileId)
	}

	rabbitmqHost, rabbitmqPort, rabbitmqUser, rabbitmqPassword := os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT"), os.Getenv("RABBITMQ_DEFAULT_USER"), os.Getenv("RABBITMQ_DEFAULT_PASS")

	rabbitmqConnector, err := connectors.CreateRabbitMQConnector(fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqUser, rabbitmqPassword, rabbitmqHost, rabbitmqPort))

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to create rabbitmq connection", Status: http.StatusInternalServerError})
		return
	}

	err = rabbitmqConnector.PublishFilesIds(context.Background(), filesIds)

	if err != nil {
		json.NewEncoder(w).Encode(Response{Message: "Unable to publish files ids", Status: http.StatusInternalServerError})
		return
	}

	json.NewEncoder(w).Encode(Response{Message: "Files uploaded", Status: http.StatusOK})
}

func serializeFile(file multipart.File, fileName string, fileSize int64, wg *sync.WaitGroup, connector *connectors.MinioConnector, ch chan FileValidationMeta) {
	defer file.Close()
	defer wg.Done()

	fileTag, err := connector.UploadFile(context.Background(), filesBucketName, file, fileName, fileSize, "application/octet-stream")

	if err != nil {
		ch <- FileValidationMeta{FileName: fileName, FileTag: "", Status: false}
		fmt.Printf("Unable to upload file %s: %s", fileName, err)
		return
	}

	ch <- FileValidationMeta{FileName: fileName, FileTag: fileTag, Status: true}
}
