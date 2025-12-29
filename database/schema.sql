CREATE TABLE assets (
    id VARCHAR(64) PRIMARY KEY,
    color VARCHAR(32),
    size INT,
    owner VARCHAR(64),
    appraised_value INT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
