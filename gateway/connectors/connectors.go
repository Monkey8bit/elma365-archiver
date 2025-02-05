package connectors

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConnector interface {
	UploadFile(ctx context.Context, bucketName string, file io.Reader, filename string, fileSize int64, contentType string) error
	MakeBucket(ctx context.Context, bucketName string) error
}

type Connector struct {
	Client *minio.Client
}

func CreateMinioConnector(endpoint, accessKey, secretKey string) (*Connector, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(strings.TrimSuffix(accessKey, "\n"), strings.TrimSuffix(secretKey, "\n"), ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return &Connector{Client: client}, nil
}

func (c *Connector) UploadFile(ctx context.Context, bucketName string, file io.Reader, filename string, fileSize int64, contentType string) (error, string) {
	fileMeta, err := c.Client.PutObject(ctx, bucketName, filename, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return err, ""
	}

	return nil, fileMeta.ETag
}
