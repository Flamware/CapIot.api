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
const mqttBroker = 'tcp://localhost:1883';
const deviceID = 'STM32-1234';
const configTopic = `config/device/${deviceID}`;
const availableTopic = `devices/available/${deviceID}`; // Device specific availability topic

// MQTT client setup
const mqttClient = mqtt.connect(mqttBroker, {
    clientId: `STM32-simulator-${deviceID}`,
    username: 'admin',
    password: 'admin',
});

mqttClient.on('connect', () => {
    logger.info(`STM32 simulator connected to MQTT broker`);
    publishAvailableMessage(); // Publish connection message to available topic
    subscribeToConfigTopic();
});

mqttClient.on('error', (error) => {
    logger.error(`STM32 simulator MQTT error:`, error);
});

function publishAvailableMessage() {
    const payload = {
        device_id: deviceID,
        timestamp: new Date().toISOString(),
    };

    mqttClient.publish(availableTopic, JSON.stringify(payload), { qos: 1 }, (error) => {
        if (error) {
            logger.error(`STM32 simulator error publishing available message:`, error);
        } else {
            logger.info(`STM32 simulator published available message:`, payload);
        }
    });
}

function subscribeToConfigTopic() {
    mqttClient.subscribe(configTopic, { qos: 1 }, (error) => {
        if (error) {
            logger.error(`STM32 simulator error subscribing to ${configTopic}:`, error);
        } else {
            logger.info(`STM32 simulator subscribed to ${configTopic}`);
        }
    });

    mqttClient.on('message', (topic, message) => {
        if (topic === configTopic) {
            try {
                const config = JSON.parse(message.toString());
                logger.info(`STM32 simulator received new configuration:`, config);
                applyConfiguration(config);
            } catch (error) {
                logger.error(`STM32 simulator error parsing configuration:`, error);
            }
        }
    });
}

function applyConfiguration(config) {
    if (config.temperatureRange) {
        logger.info(`STM32 simulator: Updating temperature range: ${JSON.stringify(config.temperatureRange)}`);
    }
    if (config.humidityRange) {
        logger.info(`STM32 simulator: Updating humidity range: ${JSON.stringify(config.humidityRange)}`);
    }
}

function sendStatusUpdate() {
    const statusTopic = `status/device/${deviceID}`;
    const statusPayload = {
        online: true,
        temperature: 25,
        humidity: 60,
    };
    mqttClient.publish(statusTopic, JSON.stringify(statusPayload), { qos: 0 }, (error) => {
        if (error) {
            logger.error(`STM32 simulator error publishing status:`, error);
        } else {
            logger.info(`STM32 simulator published status:`, statusPayload);
        }
    });
}

setInterval(sendStatusUpdate, 30000);