type ArchiverQueueItem = {
    FilesIds: number[]
};

type PostgresqlFilesSelectResponse = {
    fileNames: string[]
}

interface ServiceConnector {
    selectFiles(filesIds: number[]):  Promise<PostgresqlFilesSelectResponse>
}

export { ArchiverQueueItem, PostgresqlFilesSelectResponse, ServiceConnector };