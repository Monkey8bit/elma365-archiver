import amqp from 'amqplib';

import * as CONFIG from "../config/config";
import { MailerQueueItem } from '../types/types';

class RabbitMqConnector {
    private client: amqp.Connection | undefined;
    private archiverQueue = CONFIG.RABBITMQ.ARCHIVER_QUEUE;
    private mailerQueue = CONFIG.RABBITMQ.MAILER_QUEUE;
    channel: amqp.Channel | undefined;
    
    constructor() {
        this.client = undefined;
        this.channel = undefined;
    };

    async init() {
        const connectionUrl = `amqp://${CONFIG.RABBITMQ.USER}:${CONFIG.RABBITMQ.PASSWORD}@${CONFIG.RABBITMQ.HOST}:${CONFIG.RABBITMQ.PORT}`;

        for (let i = 0; i <= 10; i++) {
            try {
                this.client = await amqp.connect(connectionUrl);
            } catch (err) {
                await asyncDelay(2000);
                console.log(`Unable to connect to RabbitMQ from archiver" ${err}, retrying in 2 seconds..`)
            };
        };

        await this.createChannel();
    };

    async createChannel(): Promise<void> {
        if (!this.client) {
            throw new Error("Client is null, use init() method first");
        };

        this.channel = await this.client.createChannel();
        await this.channel.assertQueue(this.archiverQueue, {durable: true});
    };

    async sendMessage(message: MailerQueueItem): Promise<void> {
        const success = this.channel!.sendToQueue(this.mailerQueue, Buffer.from(JSON.stringify(message)), {contentType: "application/json"});

        if (!success) {
            console.error(`Can't send message to RabbitMQ for ${message.FileId} ${message.UserEmail}`);
        };
    };
};

async function asyncDelay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
};

export default new RabbitMqConnector();
