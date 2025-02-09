import amqp from 'amqplib/callback_api';

import * as CONFIG from './src/config/config.js';

(async () => {
    const rabbitMqConnectionUrl = `amqp://${CONFIG.RABBITMQ.USER}:${CONFIG.RABBITMQ.PASSWORD}@${CONFIG.RABBITMQ.HOST}:${CONFIG.RABBITMQ.PORT}`;
    
    amqp.connect(rabbitMqConnectionUrl, (err: Error | null, connection: amqp.Connection) => {
        if (err) {
            console.error(err);
            return;
        }
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
            channel.consume(CONFIG.RABBITMQ.ARCHIVER_QUEUE, (message) => {

                if (message !== null) {
                    const content = JSON.parse(message.content.toString());
                    console.log(content);
                    channel.ack(message);
                }
            })
        });
    });

})()