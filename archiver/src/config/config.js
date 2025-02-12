export const RABBITMQ = {
    HOST: process.env.RABBITMQ_HOST || 'localhost',
    PORT: process.env.RABBITMQ_PORT || 5672,
    USER: process.env.RABBITMQ_DEFAULT_USER || 'guest',
    PASSWORD: process.env.RABBITMQ_DEFAULT_PASS || 'guest',
    ARCHIVER_QUEUE: process.env.RABBITMQ_ARCHIVER_QUEUE || 'archiver',
    MAILER_QUEUE: process.env.RABBITMQ_MAILER_QUEUE || 'mailer',
};
export const POSTGRES = {
    HOST: process.env.POSTGRES_HOST || 'localhost',
    PORT: process.env.POSTGRES_PORT || 5432,
    USER: process.env.POSTGRES_USER || 'postgres',
    PASSWORD: process.env.POSTGRES_PASSWORD || 'postgres',
    DATABASE: process.env.POSTGRES_DB || 'elma365',
};
export const MINIO = {
    HOST: process.env.MINIO_HOST || 'localhost',
    PORT: process.env.MINIO_PORT || 9000,
    ACCESS_KEY: process.env.MINIO_ROOT_USER || 'minioadmin',
    SECRET_KEY: process.env.MINIO_ROOT_PASSWORD || 'minioadmin',
    FILES_BUCKET: process.env.MINIO_FILES_BUCKET || 'files',
    ARCHIVES_BUCKET: process.env.MINIO_ARCHIVES_BUCKET || 'archives'
};