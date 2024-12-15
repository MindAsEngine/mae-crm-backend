CREATE DATABASE IF NOT EXISTS macro_bi_cmp_528;
USE macro_bi_cmp_528;

CREATE TABLE estate_buys (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    date_added DATETIME NOT NULL,
    date_modified DATETIME NOT NULL,
    status VARCHAR(50) NOT NULL,
    status_name VARCHAR(100) NOT NULL,
    status_reason_id INT,
    INDEX idx_date_added (date_added),
    INDEX idx_status (status),
    INDEX idx_status_name (status_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE estate_buys_attributes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    entity_id BIGINT NOT NULL,
    entity VARCHAR(50) NOT NULL,
    attr_value VARCHAR(255),
    FOREIGN KEY (entity_id) REFERENCES estate_buys(id),
    INDEX idx_entity_id (entity_id),
    INDEX idx_entity (entity),
    INDEX idx_attr_value (attr_value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Grant permissions
GRANT ALL PRIVILEGES ON macro_bi_cmp_528.* TO 'user'@'%';
FLUSH PRIVILEGES;