type ArchiverQueueItem = {
    filesIds: number[]
}

type PostgresqlFilesSelectResponse = {
    s3_tags: string[]
}

interface ServiceConnector {
    selectFiles(filesIds: number[]):  Promise<PostgresqlFilesSelectResponse>
}