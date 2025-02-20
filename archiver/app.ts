import rabbitmqConnector from "./src/controllers/rabbitmq_connector";
import * as CONFIG from './src/config/config.js';
import minioConnector from './src/controllers/minio_connector';
import postgresConnector from './src/controllers/postgres_connector';
import archiveFiles from "./src/utils/archiver"

import { ArchiverQueueItem, FileForArchive, MinioObjectMeta, MailerQueueItem } from './src/types/types';

(async () => {
    console.log(`Node init, waiting for messages...`);
    await rabbitmqConnector.init();

    rabbitmqConnector.channel!.consume(CONFIG.RABBITMQ.ARCHIVER_QUEUE, async (message) => {
        if (message !== null) {
            const content: ArchiverQueueItem = JSON.parse(message.content.toString());
            let uniqueFileNames: string[] = [];

            try {
                uniqueFileNames = await postgresConnector.selectFiles(<number[]>content.FilesIds);
            } catch (err) {
                console.error(`Error at select files from postgres: ${JSON.stringify(err)}`);
                return;
            };
            
            let filesFromMinio: MinioObjectMeta[] = [];
            
            try {
                filesFromMinio = await minioConnector.getFiles(CONFIG.MINIO.FILES_BUCKET, uniqueFileNames, content.UserEmail);
            } catch (err) {
                console.error(`Error at get files from minio: ${JSON.stringify(err)}`);
            };

            let zipFileMeta: FileForArchive | undefined;

            try {
                zipFileMeta = await archiveFiles(filesFromMinio.map(file => {
                    return {
                        fileName: file!.fileName,
                        buffer: file!.buffer
                    };
                }), content.UserEmail);
            } catch (err) {
                console.error(`Error at archive operation: ${JSON.stringify(err)}`);
                return;
            };

            let eTag: string | undefined = "";
            
            try {
                eTag = await minioConnector.uploadFile(CONFIG.MINIO.ARCHIVES_BUCKET, zipFileMeta!.fileName, zipFileMeta!.buffer as Buffer, content.UserEmail);
            } catch (err) {
                console.error(`Error at uploading file ${zipFileMeta!.fileName} to minio: ${JSON.stringify(err)}`);
                return;
            };

            let fileId = 0;

            try {
                fileId = await postgresConnector.insertFile({
                    fileName: zipFileMeta!.fileName,
                    uniqueName: zipFileMeta!.fileName,
                    userId: content.UserId,
                    fileTag: eTag!
                });

                if (!fileId) {
                    throw new Error(`${eTag}: id is null`);
                };
            } catch (err) {
                console.error(`Error at insert archive in postgres: ${JSON.stringify(err)}`);
                return;
            };
            
            rabbitmqConnector.channel!.ack(message);

            await rabbitmqConnector.sendMessage({
                UserEmail: content.UserEmail,
                FileId: fileId
            });
        };
    });
})();