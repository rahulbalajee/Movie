CREATE TABLE IF NOT EXISTS movies (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    director VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS ratings (
    record_id VARCHAR(255) NOT NULL,
    record_type VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    value INT,
    PRIMARY KEY (record_id, record_type, user_id)
);
