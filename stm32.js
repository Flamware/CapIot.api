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
const availableTopic = `devices/available`; // Base topic for device availability
const configTopicBase = `config/device`; // Base topic for configuration

// Sensor configuration with individual device IDs
const sensors = [
    { id: 'airQualitySensor', type: 'airQuality', deviceID: 'AQSensor-001' },
    { id: 'specificPollutantSensor', type: 'specificPollutant', deviceID: 'PollutantSensor-002' },
    { id: 'effectivenessSensor', type: 'effectivenessMetric', deviceID: 'EffectivenessSensor-003' },
    { id: 'photocatalyseControl', type: 'control', deviceID: 'PhotocatalyseControl-004' },
    { id: 'ionisatorControl', type: 'control', deviceID: 'IonisatorControl-005' },
    { id: 'ozoneGeneratorControl', type: 'control', deviceID: 'OzoneControl-006' },
];
const sensorReadInterval = 25000; // Simulate sensor readings (though not publishing data)

// MQTT clients for each "captor"
const clients = {};

sensors.forEach(sensor => {
    const clientId = `simulator-${sensor.deviceID}`;
    const client = mqtt.connect(mqttBroker, {
        clientId: clientId,
        username: 'admin',
        password: 'admin',
    });

    client.on('connect', () => {
        logger.info(`${clientId} connected to MQTT broker`);
        publishAvailability(client, sensor.deviceID);
        subscribeToConfig(client, sensor.deviceID);
    });

    client.on('error', (error) => {
        logger.error(`${clientId} MQTT error:`, error);
    });

    clients[sensor.deviceID] = client;
});

function publishAvailability(client, deviceID) {
    const payload = {
        device_id: deviceID,
        timestamp: new Date().toISOString(),
    };
    client.publish(`${availableTopic}/${deviceID}`, JSON.stringify(payload), { qos: 1 }, (error) => {
        if (error) {
            logger.error(`Error publishing availability for ${deviceID}:`, error);
        } else {
            logger.info(`Published availability for ${deviceID}:`, payload);
        }
    });
}

function subscribeToConfig(client, deviceID) {
    const configTopic = `${configTopicBase}/${deviceID}`;
    client.subscribe(configTopic, { qos: 1 }, (error) => {
        if (error) {
            logger.error(`Error subscribing to ${configTopic} for ${deviceID}:`, error);
        } else {
            logger.info(`Subscribed to ${configTopic} for ${deviceID}`);
        }
    });

    client.on('message', (topic, message) => {
        if (topic === configTopic) {
            try {
                const config = JSON.parse(message.toString());
                logger.info(`${deviceID} received new configuration:`, config);
                applyConfiguration(deviceID, config);
            } catch (error) {
                logger.error(`Error parsing configuration for ${deviceID}:`, error);
            }
        }
    });
}

function applyConfiguration(deviceID, config) {
    switch (deviceID) {
        case 'AQSensor-001':
            if (config.samplingRate) {
                logger.info(`AQSensor-001: Setting sampling rate to ${config.samplingRate}`);
                // Apply sampling rate logic here (not actually publishing data)
            }
            break;
        case 'PollutantSensor-002':
            if (config.pollutantType) {
                logger.info(`PollutantSensor-002: Monitoring pollutant type: ${config.pollutantType}`);
                // Apply pollutant type logic here
            }
            break;
        case 'EffectivenessSensor-003':
            if (config.calibrationValue) {
                logger.info(`EffectivenessSensor-003: Setting calibration value to ${config.calibrationValue}`);
                // Apply calibration logic here
            }
            break;
        case 'PhotocatalyseControl-004':
            if (config.power) {
                logger.info(`PhotocatalyseControl-004: Setting power to ${config.power}`);
                // Apply power control logic
            }
            break;
        case 'IonisatorControl-005':
            if (config.state) {
                logger.info(`IonisatorControl-005: Setting state to ${config.state}`);
                // Apply state control logic
            }
            break;
        case 'OzoneControl-006':
            if (config.threshold) {
                logger.info(`OzoneControl-006: Setting threshold to ${config.threshold}`);
                // Apply threshold control logic
            }
            break;
        default:
            logger.warn(`Received configuration for unknown device ID: ${deviceID}`);
    }
}

function simulateSensorReadings() {
    // Simulate readings (without publishing) - this function is still running on the main process
    const airQuality = Math.floor(Math.random() * 100);
    const specificPollutant = (Math.random() * 20).toFixed(2);
    const effectivenessMetric = Math.floor(Math.random() * 100);

    // Log simulated readings for each "captor" (in the main process)
    logger.info(`Simulated Readings:`);
    logger.info(`  AQSensor-001: ${airQuality}`);
    logger.info(`  PollutantSensor-002: ${specificPollutant}`);
    logger.info(`  EffectivenessSensor-003: ${effectivenessMetric}`);
}

// Start the simulation loop (in the main process)
setInterval(simulateSensorReadings, sensorReadInterval);
logger.info(`Main simulator process running (not publishing data)`);