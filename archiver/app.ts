import amqp from 'amqplib/callback_api';

import * as CONFIG from './src/config/config.js';
import { MinioConnector } from './src/controllers/minio_controller.js';
import { postgresConnector } from './src/controllers/postgres_connector.js';
import { ArchiverQueueItem, FileForArchive } from './src/types/types.js';
import archiveFiles from "./src/utils/archiver"

(async () => {
    const rabbitMqConnectionUrl = `amqp://${CONFIG.RABBITMQ.USER}:${CONFIG.RABBITMQ.PASSWORD}@${CONFIG.RABBITMQ.HOST}:${CONFIG.RABBITMQ.PORT}`;
    const minioConnector = new MinioConnector();
    
    amqp.connect(rabbitMqConnectionUrl, (err: Error | null, connection: amqp.Connection) => {
        if (err) {
            console.error(err);
            return;
        };

        console.log(`Node init, waiting for messages...`);

        connection.createChannel((err: Error | null, channel: amqp.Channel) => {
            if (err) {
                console.error(err);
                return;
            }

            channel.assertQueue(CONFIG.RABBITMQ.ARCHIVER_QUEUE, { durable: true }, (err, ok) => {
                if (err) {
                    console.error(err);
                    return;
                }
            });
            channel.consume(CONFIG.RABBITMQ.ARCHIVER_QUEUE, async (message) => {
                if (message !== null) {
                    const content: ArchiverQueueItem = JSON.parse(message.content.toString());
                    const uniqueFileNames: string[] = await postgresConnector.selectFiles(content.FilesIds);
                    const filesFromMinio = await minioConnector.getFiles(CONFIG.MINIO.FILES_BUCKET, uniqueFileNames);
                    let zipFileMeta: FileForArchive | undefined;
                    try {
                        zipFileMeta = await archiveFiles(filesFromMinio.map(file => {
                            return {
                                fileName: file!.fileName,
                                buffer: file!.buffer
                            };
                        }), content.UserEmail);
                    } catch (err) {
                        console.error(err);
                        return;
                    };

                    let eTag: string | undefined = "";
                    
                    try {
                        eTag = await minioConnector.uploadFile(CONFIG.MINIO.ARCHIVES_BUCKET, zipFileMeta!.fileName, zipFileMeta!.buffer as Buffer);
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

                    } catch (err) {
                        console.error(`Error at insert archive in postgres: ${JSON.stringify(err)}`);
                        return;
                    };


                    console.log({filesFromMinio});
                    console.log({fileId});
                    console.log({zipFileMeta});
                    channel.ack(message);
                };
            });
        });
    });
})();