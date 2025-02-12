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
        const fileNames = await this.query(`SELECT name, unique_name FROM files WHERE id IN (${filesIds.join(",")})`);
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