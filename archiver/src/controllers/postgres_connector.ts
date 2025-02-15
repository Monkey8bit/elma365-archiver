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
        const shieldedArray = filesIds.map((_, i) => `$${++i}`).join(", ");
        const query = `SELECT unique_name FROM files WHERE id IN (${shieldedArray})`;
        const queryResult = await this.client.query<{unique_name: string}>(query, filesIds);
        return queryResult.rows.map(row => row.unique_name);
    };

    async insertFile(fileMeta: PostgresInsertMeta): Promise<number> {
        const query = "INSERT INTO files (name, s3_tag, user_id, unique_name) VALUES ($1::VARCHAR, $2::VARCHAR, $3::INT, $4::VARCHAR) RETURNING id";
        const queryResult = await this.client.query(query, [fileMeta.uniqueName, fileMeta.fileTag, fileMeta.userId, fileMeta.uniqueName]);
        
        return queryResult.rows.length > 0 ? queryResult.rows[0].id : 0;
    };
};

export default new PostgresConnector();