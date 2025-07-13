-- Create the sensor_log table
CREATE TABLE IF NOT EXISTS public.sensor_log (
                                                 log_id SERIAL PRIMARY KEY,
                                                 sensor_id VARCHAR(255) NOT NULL, -- Link to the specific sensor that generated the log
    log_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                log_content TEXT,
                                log_read BOOLEAN DEFAULT FALSE,
                                CONSTRAINT fk_sensor
                                FOREIGN KEY (sensor_id)
    REFERENCES public.sensors (sensor_id)
                            ON DELETE SET NULL -- If a sensor is deleted, logs related to it can remain, but sensor_id becomes NULL
    );