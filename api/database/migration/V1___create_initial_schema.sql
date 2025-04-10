-- db/migration/V1__create_initial_schema.sql

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) NOT NULL,
                       name VARCHAR(255),
                       password VARCHAR(255),
                       role VARCHAR(50),
                       auth0_id VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE devices (
                         device_id TEXT PRIMARY KEY,
                         timestamp TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE locations (
                           location_id SERIAL PRIMARY KEY,
                           location_name TEXT NOT NULL,
                           location_description TEXT
);

CREATE TABLE user_location (
                               user_id INTEGER NOT NULL,
                               location_id INTEGER NOT NULL,
                               assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (user_id, location_id),
                               CONSTRAINT user_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES locations(location_id) ON DELETE CASCADE,
                               CONSTRAINT user_location_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE device_location (
                                 id SERIAL PRIMARY KEY,
                                 device_id TEXT,
                                 location_id INTEGER,
                                 assigned_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
                                 is_current BOOLEAN DEFAULT true,
                                 CONSTRAINT device_location_device_id_location_id_key UNIQUE (device_id, location_id),
                                 CONSTRAINT unique_device_location UNIQUE (device_id, location_id),
                                 CONSTRAINT device_location_device_id_fkey FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE,
                                 CONSTRAINT device_location_location_id_fkey FOREIGN KEY (location_id) REFERENCES locations(location_id) ON DELETE CASCADE
);

-- Create the unique index with WHERE clause AFTER the table is created
CREATE UNIQUE INDEX unique_current_device_location ON device_location (device_id) WHERE is_current = true;

-- Triggers for device_location (as discussed previously, you'll need the function definitions)
-- CREATE FUNCTION update_current_location() RETURNS TRIGGER AS $$ ... $$ LANGUAGE plpgsql;
-- CREATE TRIGGER set_current_location AFTER INSERT OR UPDATE ON device_location ...;
-- CREATE FUNCTION update_device_location() RETURNS TRIGGER AS $$ ... $$ LANGUAGE plpgsql;
-- CREATE TRIGGER set_device_current_location BEFORE INSERT OR UPDATE ON device_location ...;
-- CREATE TRIGGER update_current_location_trigger AFTER INSERT ON device_location ...;