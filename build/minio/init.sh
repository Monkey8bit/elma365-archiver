#!/bin/sh
sleep 5

mc alias set minio http://minio:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

until mc ls minio; do
    echo "Waiting for MinIO..."
    sleep 2
done

if ! mc ls minio/files &> /dev/null; then
    mc mb minio/files
fi

if ! mc ls minio/archives &> /dev/null; then
    mc mb minio/archives
fi