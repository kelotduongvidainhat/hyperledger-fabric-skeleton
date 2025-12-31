CREATE TABLE assets (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(64),
    category VARCHAR(64),
    owner VARCHAR(64),
    status VARCHAR(32),
    updated VARCHAR(64),
    updated_by VARCHAR(64),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(64),
    role VARCHAR(32),
    status VARCHAR(32),
    updated VARCHAR(64),
    updated_by VARCHAR(64),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
