package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnector struct {
	client *amqp091.Connection
}

type PostgresConnector struct {
	pgpool *pgxpool.Pool
}

type MinioConnector struct {
	connection *minio.Client
}

type MailerQueueItem struct {
	FileId    int
	UserEmail string
}

var MINIO_PROXY_HOST = os.Getenv("MINIO_PROXY_HOST")
var MINIO_PROXY_PORT = os.Getenv("MINIO_PROXY_PORT")

func CreateRabbitMQConnector(connString string) (*RabbitMQConnector, error) {
	var err error
	var connection *amqp091.Connection

	for i := 0; i < 10; i++ {
		connection, err = amqp091.Dial(connString)

		if err == nil {
			return &RabbitMQConnector{client: connection}, nil
		}

		log.Printf("Unable to connect to RabbitMQ: %s, retrying in 2 seconds..", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func (c *RabbitMQConnector) CreateChannel() (*amqp091.Channel, error) {
	return c.client.Channel()
}

func (c *RabbitMQConnector) Consume(queue string, channel *amqp091.Channel, ch chan *amqp091.Delivery) error {
	messages, err := channel.Consume(
		queue,
		"mailer",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	log.Println("consumer created")

	go func() {
		for message := range messages {
			ch <- &message
		}
	}()

	return nil
}

func (c *RabbitMQConnector) HandleRabbitMessage(message *amqp091.Delivery) (*MailerQueueItem, error) {
	var parcedMessage MailerQueueItem

	err := json.Unmarshal(message.Body, &parcedMessage)

	if err != nil {
		log.Panicf("Error on parse message: %s", message.MessageId)
		return nil, err
	}

	return &parcedMessage, nil
}
func (c *RabbitMQConnector) Close() error {
	return c.client.Close()
}

func CreatePool(connString string) (*PostgresConnector, error) {
	var err error
	for i := 0; i < 10; i++ {
		conn, err := pgxpool.New(context.Background(), connString)

		if err == nil {
			return &PostgresConnector{pgpool: conn}, nil
		}

		log.Printf("Unable to create Postgresql pool: %s, retrying in 2 seconds..", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func (p *PostgresConnector) GetUniqueName(ctx context.Context, fileId int) (string, error) {
	var uniqueName string

	err := p.pgpool.QueryRow(ctx, "SELECT unique_name FROM files WHERE id=$1", fileId).Scan(&uniqueName)

	if err != nil {
		log.Printf("Error at select: %s", err)
		return uniqueName, err
	}

	return uniqueName, nil
}

func CreateMinioConnector(endpoint, accessKey, secretKey string) (*MinioConnector, error) {
	var err error

	for i := 0; i < 10; i++ {
		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})

		if err == nil {
			return &MinioConnector{connection: client}, nil
		}

		log.Printf("Unable to connect to Minio: %s, retrying in 2 seconds..", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func (c MinioConnector) GetFileDownloadLink(uniqueName string, ctx context.Context, bucketName string, objectName string) (string, error) {
	log.Println(uniqueName, bucketName, objectName)
	fileLink, err := c.connection.PresignedGetObject(ctx, bucketName, objectName, time.Hour*1, url.Values{})

	if err != nil {
		log.Panicf("Error at get file from minio: %s", err)
		return "", err
	}

	stringFileLink := fileLink.String()
	rewritedFileLink := strings.Replace(stringFileLink, c.connection.EndpointURL().Host, fmt.Sprintf("%s:%s", MINIO_PROXY_HOST, MINIO_PROXY_PORT), 1)

	return rewritedFileLink, nil
}
