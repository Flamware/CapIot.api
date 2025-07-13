-- db/migration/V1__create_initial_schema_and_sensors.sql

-- Create the users table
CREATE TABLE IF NOT EXISTS public.users (
                                            id SERIAL PRIMARY KEY,
                                            email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    password VARCHAR(255),
    role VARCHAR(50),
    auth0_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Create the devices table
CREATE TABLE IF NOT EXISTS public.devices (
                                              device_id TEXT PRIMARY KEY,
                                              last_seen TIMESTAMP WITH TIME ZONE,
                                              status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Create the locations table
CREATE TABLE IF NOT EXISTS public.locations (
                                                location_id SERIAL PRIMARY KEY,
                                                location_name TEXT NOT NULL,
                                                location_description TEXT
);

-- Create the user_location table
CREATE TABLE IF NOT EXISTS public.user_location (
                                                    user_id INTEGER NOT NULL,
                                                    location_id INTEGER NOT NULL,
                                                    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                    PRIMARY KEY (user_id, location_id),
    CONSTRAINT user_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON DELETE CASCADE,
    CONSTRAINT user_location_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
    );

-- Create the device_location table
CREATE TABLE IF NOT EXISTS public.device_location (
                                                      id SERIAL PRIMARY KEY,
                                                      device_id TEXT,
                                                      location_id INTEGER,
                                                      assigned_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    is_current BOOLEAN DEFAULT true,
    CONSTRAINT device_location_device_id_location_id_key UNIQUE (device_id, location_id),
    CONSTRAINT unique_device_location UNIQUE (device_id, location_id),
    CONSTRAINT device_location_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE,
    CONSTRAINT device_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(location_id) ON DELETE CASCADE
    );

-- Create the unique index with WHERE clause AFTER the table is created
CREATE UNIQUE INDEX IF NOT EXISTS unique_current_device_location ON public.device_location (device_id) WHERE is_current = true;

-- Create the sensors table
CREATE TABLE IF NOT EXISTS public.sensors (
                                              sensor_id VARCHAR(255) PRIMARY KEY,
    sensor_type VARCHAR(255) NOT NULL
    );

-- Create the device_sensors table
CREATE TABLE IF NOT EXISTS public.device_sensors (
                                                     device_id TEXT NOT NULL,
                                                     sensor_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (device_id, sensor_id),
    FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE,
    FOREIGN KEY (sensor_id) REFERENCES public.sensors(sensor_id) ON DELETE CASCADE
    );

-- Grant privileges to the 'admin' role on all tables in the public schema
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO admin;

-- Grant USAGE privilege to the 'admin' role on all sequences in the public schema
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO admin;