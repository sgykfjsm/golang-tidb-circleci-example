CREATE DATABASE IF NOT EXISTS test;
USE test;
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,
    name VARCHAR(64) NOT NULL
);
-- EOF
