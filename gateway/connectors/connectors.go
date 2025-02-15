package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rabbitmq/amqp091-go"
)

type MinioConnector struct {
	Client *minio.Client
}

type PostgresConnector struct {
	Client *pgx.Conn
}

type RabbitMQConnector struct {
	Client *amqp091.Connection
}

type FileMeta struct {
	FileName   string
	MinioTag   string
	UserId     int
	UniqueName string
}

type RabbitMQMessage struct {
	FilesIds  []int
	UserEmail string
	UserId    int
}

func CreateMinioConnector(endpoint, accessKey, secretKey string) (*MinioConnector, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(strings.TrimSuffix(accessKey, "\n"), strings.TrimSuffix(secretKey, "\n"), ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return &MinioConnector{Client: client}, nil
}

func (c *MinioConnector) UploadFile(ctx context.Context, bucketName string, file io.Reader, name string, fileSize int64, contentType string, userEmail string) (string, string, error) {
	fileNameArray := strings.Split(name, ".")
	uniqueName := ""

	if len(fileNameArray) > 1 {
		fileName, fileExtension := strings.Join(fileNameArray[:len(fileNameArray)-1], "."), fileNameArray[len(fileNameArray)-1]
		uniqueName = fmt.Sprintf("%s_%s.%s", fileName, uuid.New().String(), fileExtension)
	} else {
		uniqueName = fmt.Sprintf("%s_%s", name, uuid.New().String())
	}

	fileMeta, err := c.Client.PutObject(ctx, bucketName, fmt.Sprintf("%s/%s", userEmail, uniqueName), file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"file_name": name,
		},
	})

	if err != nil {
		return "", "", err
	}

	return fileMeta.ETag, uniqueName, nil
}

func CreatePostgresConnector(connString string) (*PostgresConnector, error) {
	conn, err := pgx.Connect(context.Background(), connString)

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return &PostgresConnector{Client: conn}, nil
}

func (c *PostgresConnector) GetOrCreateUser(ctx context.Context, email string) (int, error) {
	var userId int
	err := c.Client.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", email).Scan(&userId)

	if userId == 0 {
		log.Println("User already exists")
		err = c.Client.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userId)
	}

	if err != nil {
		log.Println(err)
		return userId, err
	}

	return userId, nil

}

func (c *PostgresConnector) InsertFile(ctx context.Context, fileMeta FileMeta) (int, error) {
	var err error
	var fileId int

	err = c.Client.QueryRow(ctx, "INSERT INTO files (name, s3_tag, user_id, unique_name) VALUES ($1, $2, $3, $4) RETURNING id", fileMeta.FileName, fileMeta.MinioTag, fileMeta.UserId, fileMeta.UniqueName).Scan(&fileId)

	if err != nil {
		log.Println(err)
		return fileId, err
	}

	return fileId, nil
}

func CreateRabbitMQConnector(connString string) (*RabbitMQConnector, error) {
	conn, err := amqp091.Dial(connString)

	if err != nil {
		log.Print("Unable to connect to RabbitMQ")
		log.Fatalln(err)
		return nil, err
	}

	return &RabbitMQConnector{Client: conn}, nil
}

func (c *RabbitMQConnector) PublishFilesIds(ctx context.Context, filesIds []int, userEmail string, userId int) error {
	ch, err := c.Client.Channel()

	if err != nil {
		log.Println(err)
		return err
	}

	defer ch.Close()

	body, err := json.Marshal(RabbitMQMessage{FilesIds: filesIds, UserEmail: userEmail, UserId: userId})

	if err != nil {
		log.Println(err)
		return err
	}

	archiverQueueName := os.Getenv("RABBITMQ_ARCHIVER_QUEUE")

	err = ch.Publish(
		"",
		archiverQueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2,
		},
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
