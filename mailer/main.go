package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"elma365-archiver/mailer/connectors"
	"elma365-archiver/mailer/utils"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmqHost, rabbitmqPort, rabbitmqUser, rabbitmqPassword := os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT"), os.Getenv("RABBITMQ_DEFAULT_USER"), os.Getenv("RABBITMQ_DEFAULT_PASS")
	rabbitmqConnector, err := connectors.CreateRabbitMQConnector(fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqUser, rabbitmqPassword, rabbitmqHost, rabbitmqPort))

	if err != nil {
		log.Printf("Error on RabbitMQ connect: %s", err)
	}

	defer rabbitmqConnector.Close()

	rabbitmqQueueName := os.Getenv("RABBITMQ_MAILER_QUEUE")
	rabbitChannel, err := rabbitmqConnector.CreateChannel()

	if err != nil {
		log.Printf("Error on create RabbitMQ channel: %s", err)
		return
	}

	defer rabbitChannel.Close()

	ctx := context.Background()
	postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB")
	postgresConnector, err := connectors.CreatePool(fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", postgresUser, postgresPassword, postgresHost, postgresPort, postgresDb))

	if err != nil {
		log.Printf("Error on create Postgresql pool: %s", err)
		return
	}

	accessKey, secretKey, minioHost, minioPort, bucketName := os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), os.Getenv("MINIO_HOST"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_ARCHIVES_BUCKET")

	minioConnector, err := connectors.CreateMinioConnector(fmt.Sprintf("%s:%s", minioHost, minioPort), accessKey, secretKey)

	if err != nil {
		log.Printf("Error at create minio conector: %s", err)
		return
	}

	smtpConnector, err := utils.CreateSMTPConnector()

	if err != nil {
		log.Printf("error at CreateSMTPConnector: %s", err)
		return
	}

	messageChannel := make(chan *amqp091.Delivery)

	err = rabbitmqConnector.Consume(rabbitmqQueueName, rabbitChannel, messageChannel)

	go func() {
		for msg := range messageChannel {
			if err != nil {
				log.Printf("Error while consume: %s", err)
				continue
			}
			parcedMessage, err := rabbitmqConnector.HandleRabbitMessage(msg)

			if err != nil {
				log.Printf("Error at HandleRabbitMessage: %s", err)
				continue
			}

			uniqueName, err := postgresConnector.GetUniqueName(ctx, parcedMessage.FileId)

			if err != nil {
				log.Printf("Error at GetUniqueName: %s", err)
				continue
			}

			fileLink, err := minioConnector.GetFileDownloadLink(uniqueName, ctx, bucketName, fmt.Sprintf("%s/%s", parcedMessage.UserEmail, uniqueName))

			if err != nil {
				log.Printf("Error at GetUniqueName: %s", err)
				continue
			}

			err = smtpConnector.SendMail(ctx, parcedMessage.UserEmail, fileLink)

			if err != nil {
				log.Printf("Error at SendMail: %s", err)
				continue
			}
		}
	}()

	if err != nil {
		log.Println(err)
	}

	select {}
}
