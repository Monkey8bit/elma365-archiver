import { Readable } from "node:stream";

type ArchiverQueueItem = {
    FilesIds: number[] | number,
    UserEmail: string,
    UserId: number
};

type MailerQueueItem = {
    FileId: number,
    UserEmail: string,
}

type PostgresqlFilesSelectResponse = {
    fileNames: string[]
};

type PostgresInsertMeta = {
    userId: number,
    fileTag: string,
    fileName: string,
    uniqueName: string
};


type MinioObjectMeta = {
    fileName: string,
    uniqueName: string,
    buffer: Readable,
};

type FileForArchive = {
    fileName: string,
    buffer: Readable | Buffer
};


export { ArchiverQueueItem, PostgresqlFilesSelectResponse, MinioObjectMeta, Readable, FileForArchive, PostgresInsertMeta, MailerQueueItem };