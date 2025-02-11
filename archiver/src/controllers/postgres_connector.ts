import * as CONFIG from "../config/config.js";
import pg from "pg";

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

    async selectFiles(filesIds: number[]) {
        console.log({filesIds});
        const fileNames = await this.query(`SELECT name FROM files WHERE id IN (${filesIds.join(",")})`);
        console.log(fileNames);
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