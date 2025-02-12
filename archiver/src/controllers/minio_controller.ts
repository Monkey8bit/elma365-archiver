import { Client } from "minio";
import * as CONFIG from "../config/config.js"
import { MinioObjectCacheItem } from "../types/types";
import { BucketItemStat } from "minio";

const CHUNK_SIZE = 1000;

class MinioConnector {
    private client: Client;
    private cache: MinioObjectCacheItem[] = [];

    constructor() {
        console.log("Minio connector init");

        this.client = new Client({
            endPoint: CONFIG.MINIO.HOST,
            port: Number(CONFIG.MINIO.PORT),
            accessKey: CONFIG.MINIO.ACCESS_KEY,
            secretKey: CONFIG.MINIO.SECRET_KEY,
            useSSL: false
        });
        this.listObjects(CONFIG.MINIO.FILES_BUCKET).then(() => {
            console.log(`Minio cache init, ${this.cache.length} objects cached: ${JSON.stringify(this.cache)}`);
        });
    };

    private async getObjMeta(fileName: string, obj: BucketItemStat) {
        return {
            fileName,
            uniqueName: obj.metaData['unique_name']
        };
    };

    private async listObjects(bucketName: string) {
        const stream = this.client.listObjectsV2(bucketName, '', true);
        let chunk: Promise<MinioObjectCacheItem>[] = [];

        for await (let obj of stream) {
            chunk.push(this.client.statObject(bucketName, obj.name).then(stat => {
                return this.getObjMeta(obj.name, stat);
            }));

            if (chunk.length >= CHUNK_SIZE) {
                this.cache.push(...await Promise.all(chunk));
                chunk = [];
            };
        };

        this.cache.push(...await Promise.all(chunk));
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
        this.client.listObjectsV2
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

export { MinioConnector };