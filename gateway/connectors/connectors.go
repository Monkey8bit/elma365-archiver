package connectors

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

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
	FileName string
	MinioTag string
	UserId   int
}

type RabbitMQMessage struct {
	FilesIds []int
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

func (c *MinioConnector) UploadFile(ctx context.Context, bucketName string, file io.Reader, filename string, fileSize int64, contentType string) (string, error) {
	fileTag, err := c.checkFileExists(ctx, bucketName, filename)

	if err != nil {
		return "", err
	}

	if len(fileTag) > 0 {
		return fileTag, nil
	}

	fileMeta, err := c.Client.PutObject(ctx, bucketName, filename, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return "", err
	}

	return fileMeta.ETag, nil
}

func (c *MinioConnector) checkFileExists(ctx context.Context, bucketName string, fileName string) (string, error) {
	fileMeta, err := c.Client.StatObject(ctx, bucketName, fileName, minio.StatObjectOptions{})

	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return "", nil
		}

		return "", err
	}

	return fileMeta.ETag, nil
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

	err = c.Client.QueryRow(ctx, "SELECT id FROM files WHERE s3_tag = $1 AND name = $2", fileMeta.MinioTag, fileMeta.FileName).Scan(&fileId)

	if err != nil && err != pgx.ErrNoRows {
		log.Println(err)
		return fileId, err
	}

	if fileId != 0 {
		log.Printf("File %s with id %d already exists\n", fileMeta.FileName, fileId)
		return fileId, nil
	}

	err = c.Client.QueryRow(ctx, "INSERT INTO files (name, s3_tag, user_id) VALUES ($1, $2, $3) RETURNING id", fileMeta.FileName, fileMeta.MinioTag, fileMeta.UserId).Scan(&fileId)

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

func (c *RabbitMQConnector) PublishFilesIds(ctx context.Context, filesIds []int) error {
	ch, err := c.Client.Channel()

	if err != nil {
		log.Println(err)
		return err
	}

	defer ch.Close()

	body, err := json.Marshal(RabbitMQMessage{FilesIds: filesIds})

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
