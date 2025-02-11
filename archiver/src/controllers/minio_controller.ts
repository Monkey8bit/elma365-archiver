import minio from "minio";
import * as CONFIG from "../config/config.js"

class MinioConnector {
    private client: minio.Client;

    constructor() {
        this.client = new minio.Client({
            endPoint: CONFIG.MINIO.HOST,
            port: Number(CONFIG.MINIO.PORT),
            accessKey: CONFIG.MINIO.ACCESS_KEY,
            secretKey: CONFIG.MINIO.SECRET_KEY,
        });
    };
    
    private async getFile(bucketName: string, fileName: string) {
        return this.client.getObject(bucketName, fileName).then(res => {
            return {
                fileName,
                buffer: res
            };
        }).catch(err => {
            return {
                fileName,
                buffer: null,
                error: err.message
            }
        });
    };

    async getFiles(bucketName: string, minioFilesNames: string[]) {
        const files = await Promise.all(minioFilesNames.map(fileName => this.getFile(bucketName, fileName)));

        files.filter(file => file.buffer === null).forEach(file => {
            console.error(`Error while fetching file ${file.fileName}: ${file.error}`);
        });

        return files.filter(file => file.buffer !== null);
    };

    async uploadFile(bucketName: string, fileName: string, fileBuffer: Buffer) {
        return this.client.putObject(bucketName, fileName, fileBuffer).then(res => {
            return res.etag;
        }).catch(err => {
            throw new Error(err.message);
        });
    };
};

export const minioConnector = new MinioConnector();