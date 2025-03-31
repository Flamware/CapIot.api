const mqtt = require('mqtt');
const winston = require('winston');

// Logging setup
const logger = winston.createLogger({
    level: 'info',
    format: winston.format.json(),
    transports: [
        new winston.transports.Console({ format: winston.format.simple() }),
    ],
});

// MQTT broker details
const mqttBroker = 'tcp://localhost:1883'; // Change if needed
const deviceID = 'STM32-1234'; // Replace with the target device ID
const configTopic = `config/device/${deviceID}`;

// MQTT client setup
const mqttClient = mqtt.connect(mqttBroker, {
    clientId: 'config-publisher', // Unique client ID
    // Add username and password if using authentication
    username: 'admin',
    password: 'admin',
});

mqttClient.on('connect', () => {
    logger.info('Configuration publisher connected to MQTT broker');
    publishNewConfig(); // Publish initial config
    setInterval(publishNewConfig, 20000); // Publish every 20 seconds
});

mqttClient.on('error', (error) => {
    logger.error('Configuration publisher MQTT error:', error);
});

function publishNewConfig() {
    const newConfig = {
        temperatureRange: {
            min: Math.floor(Math.random() * 10), // Random min temperature
            max: Math.floor(Math.random() * 20) + 30, // Random max temperature
        },
        humidityRange: {
            min: Math.floor(Math.random() * 20), // Random min humidity
            max: Math.floor(Math.random() * 40) + 60, // Random max humidity
        },
        // Add more configuration parameters as needed
    };

    const payload = JSON.stringify(newConfig);

    mqttClient.publish(configTopic, payload, { qos: 1 }, (error) => {
        if (error) {
            logger.error(`Configuration publisher error publishing to ${configTopic}:`, error);
        } else {
            logger.info(`Configuration publisher published new config:`, newConfig);
        }
    });
}