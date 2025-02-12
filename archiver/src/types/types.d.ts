type ArchiverQueueItem = {
    FilesIds: number[]
};

type PostgresqlFilesSelectResponse = {
    fileNames: string[]
}

type MinioObjectCacheItem = {
    fileName: string,
    uniqueName: string
}

interface ServiceConnector {
    selectFiles(filesIds: number[]):  Promise<PostgresqlFilesSelectResponse>
}

export { ArchiverQueueItem, PostgresqlFilesSelectResponse, ServiceConnector, MinioObjectCacheItem };