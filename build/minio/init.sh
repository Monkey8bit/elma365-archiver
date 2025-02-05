#!/bin/sh
mc alias set minio http://minio:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

echo $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

until mc ls minio; do
    echo "Waiting for MinIO..."
    sleep 2
done

mc mb minio/files && mc mb minio/archives