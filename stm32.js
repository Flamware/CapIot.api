const mqtt = require('mqtt');
const winston = require('winston');
const moment = require('moment');

// Configuration du logger
const logger = winston.createLogger({
    level: 'info',
    format: winston.format.json(),
    transports: [
        new winston.transports.Console({ format: winston.format.simple() }),
    ],
});

// Configuration MQTT
const mqttBroker = 'tcp://localhost:1883';
const deviceID = 'STM32-Simulator-001'; // ID unique pour ce simulateur
const availabilityTopic = `devices/available/${deviceID}`;
const configTopic = `config/device/${deviceID}`;
const configAckTopic = `devices/${deviceID}/config/ack`;
const statusTopic = `devices/status/${deviceID}`;
const heartbeatTopic = `devices/heartbeat/${deviceID}`; // Topic dédié au heartbeat
const dataTopicBase = `iot/data/${deviceID}`;
const startCaptorsCommandTopic = `devices/${deviceID}/command/start_captors`;

// Configuration des capteurs simulés
const sensors = [
    { type: 'temperature', id: 'temp-sim-001', currentValue: 25.5 },
    { type: 'humidity', id: 'hum-sim-001', currentValue: 60.2 },
    { type: 'pressure', id: 'press-sim-001', currentValue: 1012.3 },
];

let isConfigured = false;
let isMonitoring = false;
let currentConfig = {};

// Options de connexion MQTT avec LWT
const connectOptions = {
    clientId: deviceID,
    username: 'admin', // Si votre broker requiert une authentification
    password: 'admin',
    will: {
        topic: `devices/lwt/${deviceID}`, // Changed topic for LWT
        payload: JSON.stringify({ device_id: deviceID, status: 'offline', timestamp: moment().toISOString() }),
        qos: 1,
        retain: false,
    },
};

// Création du client MQTT
const client = mqtt.connect(mqttBroker, connectOptions);

client.on('connect', () => {
    logger.info(`${deviceID} connecté au broker MQTT`);

    // Publication immédiate de l'availability
    publishAvailability();

    // Abonnement aux topics de configuration et de démarrage des capteurs
    client.subscribe(configTopic, { qos: 1 });
    client.subscribe(startCaptorsCommandTopic, { qos: 1 });

    // Publication périodique du statut "running" (peut être utilisé comme un heartbeat de base)
    setInterval(publishOnlineStatus, 6000); // Toutes les 60 secondes

    // Publication périodique du heartbeat sur un topic dédié
    setInterval(publishHeartbeat, 3000); // Toutes les 30 secondes (plus fréquent que le statut)
});

client.on('message', (topic, message) => {
    const payloadString = message.toString();
    logger.info(`${deviceID} a reçu un message sur le topic ${topic}: ${payloadString}`);

    try {
        const payload = JSON.parse(payloadString);

        if (topic === configTopic) {
            handleConfiguration(payload);
        } else if (topic === startCaptorsCommandTopic) {
            handleStartCaptorsCommand();
        }
    } catch (error) {
        logger.error(`${deviceID} Erreur lors du parsing du message sur ${topic}: ${error}`);
    }
});

client.on('error', (error) => {
    logger.error(`${deviceID} Erreur MQTT: ${error}`);
});

client.on('disconnect', () => {
    logger.info(`${deviceID} déconnecté du broker MQTT`);
});

function publishAvailability() {
    const capabilities = sensors.map(sensor => ({
        captor_type: sensor.type,
        captor_id: sensor.id,
    }));

    const payload = {
        device_id: deviceID,
        status: 'online',
        timestamp: moment().toISOString(),
        captors: capabilities,
    };

    client.publish(availabilityTopic, JSON.stringify(payload), { qos: 1, retain: false }, (error) => {
        if (error) {
            logger.error(`${deviceID} Erreur lors de la publication de l'availability: ${error}`);
        } else {
            logger.info(`${deviceID} Availability publiée sur ${availabilityTopic}: ${JSON.stringify(payload)}`);
        }
    });
}

function publishOnlineStatus() {
    const payload = {
        device_id: deviceID,
        status: 'running',
        timestamp: moment().toISOString(),
    };

    client.publish(statusTopic, JSON.stringify(payload), { qos: 1, retain: false }, (error) => {
        if (error) {
            logger.error(`${deviceID} Erreur lors de la publication du statut en ligne: ${error}`);
        } else {
            logger.info(`${deviceID} Statut "running" publié sur ${statusTopic}: ${JSON.stringify(payload)}`);
        }
    });
}

function publishHeartbeat() {
    const payload = {
        device_id: deviceID,
        timestamp: moment().toISOString(),
    };

    client.publish(heartbeatTopic, JSON.stringify(payload), { qos: 0, retain: false }, (error) => {
        if (error) {
            logger.error(`${deviceID} Erreur lors de la publication du heartbeat: ${error}`);
        } else {
            logger.info(`${deviceID} Heartbeat publié sur ${heartbeatTopic}: ${JSON.stringify(payload)}`);
        }
    });
}

function handleConfiguration(config) {
    logger.info(`${deviceID} a reçu une configuration: ${JSON.stringify(config)}`);
    currentConfig = config;
    isConfigured = true;

    // Simuler l'application de la configuration (vous feriez ici la logique réelle)
    logger.info(`${deviceID} Configuration appliquée.`);

    // Envoyer un acknowledgement de configuration
    const ackPayload = {
        status: 'configured',
        timestamp: moment().toISOString(),
    };
    client.publish(configAckTopic, JSON.stringify(ackPayload), { qos: 1 }, (error) => {
        if (error) {
            logger.error(`${deviceID} Erreur lors de l'envoi de l'ACK de configuration: ${error}`);
        } else {
            logger.info(`${deviceID} ACK de configuration publié sur ${configAckTopic}: ${JSON.stringify(ackPayload)}`);
        }
    });
}

function handleStartCaptorsCommand() {
    if (isConfigured) {
        logger.info(`${deviceID} a reçu la commande de démarrage des capteurs.`);
        isMonitoring = true;
        startDataMonitoring();
    } else {
        logger.warn(`${deviceID} a reçu la commande de démarrage avant d'être configuré.`);
    }
}

function startDataMonitoring() {
    if (isMonitoring) {
        logger.info(`${deviceID} Démarrage de la surveillance et de la publication des données.`);
        setInterval(publishSensorData, 5000); // Publier les données toutes les 5 secondes
    }
}

function publishSensorData() {
    if (isMonitoring) {
        const timestamp = moment().toISOString();
        const data = {};
        sensors.forEach(sensor => {
            // Simuler la lecture du capteur (ajouter un peu de variation)
            sensor.currentValue += (Math.random() - 0.5) * 0.1;
            data[sensor.type] = sensor.currentValue.toFixed(2);
            const payload = {
                device_id: deviceID,
                sensor_id: sensor.id,
                type: sensor.type,
                value: sensor.currentValue.toFixed(2),
                timestamp: timestamp,
            };
            const dataTopic = `${dataTopicBase}/${sensor.type}`;
            client.publish(dataTopic, JSON.stringify(payload), { qos: 0 }, (error) => {
                if (error) {
                    logger.error(`${deviceID} Erreur lors de la publication des données de ${sensor.type}: ${error}`);
                } else {
                    logger.info(`${deviceID} Données de ${sensor.type} publiées sur ${dataTopic}: ${JSON.stringify(payload)}`);
                }
            });
        });
    }
}