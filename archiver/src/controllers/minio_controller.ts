import { Client } from "minio";
import * as CONFIG from "../config/config.js";
import { MinioObjectMeta } from "../types/types";
import crypto from "crypto";

const CHUNK_SIZE = 1000;

class MinioConnector {
    private client: Client;

    constructor() {
        console.log("Minio connector init");

        this.client = new Client({
            endPoint: CONFIG.MINIO.HOST,
            port: Number(CONFIG.MINIO.PORT),
            accessKey: CONFIG.MINIO.ACCESS_KEY,
            secretKey: CONFIG.MINIO.SECRET_KEY,
            useSSL: false
        });
    };

    private async getFile(bucketName: string, fileName: string): Promise<MinioObjectMeta | undefined> {
        function isMinioError(error: unknown): error is {code: string, message: string} {
            return !!error && typeof error === "object" && "code" in error;
        };

        try {
            const objStat = await this.client.statObject(bucketName, fileName);
            
            return this.client.getObject(bucketName, fileName).then(res => {
                return {
                    uniqueName: fileName,
                    buffer: res,
                    fileName: objStat.metaData["file_name"],
                };
            }).catch(err => {
                console.error(`Error getting file ${fileName} from minio: ${err.message}`);
                return undefined;
            });
        } catch (err: unknown) {
            if (isMinioError(err) && err.code === "NotFound") {
                console.error(`File with name ${fileName} doesn't exists`);
                return;
            } else {
                console.log(err)
            };
        };
    };

    async getFiles(bucketName: string, minioFilesNames: string[]) {
        const files = await Promise.all(minioFilesNames.map(fileName => this.getFile(bucketName, fileName))).then(data => data.filter(Boolean));
        console.log({files})
        return files.filter(file => file!.buffer !== null);
    };
    
    async uploadFile(bucketName: string, name: string, fileBuffer: Buffer): Promise<string | undefined> {
        const fileNameArr = name.split(".");
        let uniqueName = "";

        if (fileNameArr.length > 1) {
            const [fileName, fileExtention] = [fileNameArr.slice(0, fileNameArr.length - 1), fileNameArr[fileNameArr.length - 1]];
            uniqueName = `${fileName}.${fileExtention}`;
        } else {
            uniqueName = name;
        };

        return this.client.putObject(bucketName, uniqueName, fileBuffer, fileBuffer.length, {file_name: name}).then(res => {
            return res.etag;
        }).catch(err => {
            console.error(err);
            return undefined;
        });
    };
};

export { MinioConnector };