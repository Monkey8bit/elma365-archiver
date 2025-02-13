import * as CONFIG from "../config/config.js";
import pg from "pg";
import { PostgresInsertMeta } from "../types/types"

class PostgresConnector {
    private client: pg.Pool;

    constructor() {
        this.client = new pg.Pool({
            user: CONFIG.POSTGRES.USER,
            host: CONFIG.POSTGRES.HOST,
            database: CONFIG.POSTGRES.DATABASE,
            password: CONFIG.POSTGRES.PASSWORD,
            port: Number(CONFIG.POSTGRES.PORT),
        });
    };

    async selectFiles(filesIds: number[]): Promise<string[]> {
        const fileNames = await this.query(`SELECT unique_name FROM files WHERE id IN (${filesIds.join(",")})`);
        return fileNames.map(obj => obj.unique_name);
    };

    async insertFile(fileMeta: PostgresInsertMeta): Promise<number> {
        const queryPromise = await this.query(`INSERT INTO archives (name, s3_tag, user_id, unique_name) VALUES (${Object.keys(fileMeta).map(key => {
            return `'${fileMeta[key as keyof typeof fileMeta]}'`
        }).join(", ")}) RETURNING id`);
        
        return queryPromise.length > 0 ? queryPromise[0].id : 0;
    };

    private async query(query: string) {
        return this.client.query(query).then(res => {
            return res.rows;
        }).catch(err => {
            throw new Error(err.message);
        });
    };
};

export const postgresConnector = new PostgresConnector();