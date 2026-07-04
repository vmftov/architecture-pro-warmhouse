CREATE DATABASE telemetry_db;

\c telemetry_db;

CREATE TABLE sensor_telemetry (
    sensor_id SERIAL NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    value FLOAT NOT NULL DEFAULT 0,

    PRIMARY KEY (sensor_id)
);

CREATE INDEX idx_sensor_telemetry_timestamp ON sensor_telemetry(timestamp);

CREATE TABLE relay_values (
    relay_id SERIAL NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    value FLOAT NOT NULL DEFAULT 0,

    PRIMARY KEY (relay_id)
);

CREATE INDEX idx_relay_values_timestamp ON relay_values(timestamp);

