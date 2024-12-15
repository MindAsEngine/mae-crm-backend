CREATE DATABASE reports_db;

CREATE TABLE audiences (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audience_requests (
    id SERIAL PRIMARY KEY,
    audience_id INTEGER REFERENCES audiences(id),
    request_id INTEGER NOT NULL,
    status VARCHAR(50),
    reason VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE audience_integrations (
    id SERIAL PRIMARY KEY,
    audience_id INTEGER REFERENCES audiences(id),
    cabinet_id INTEGER NOT NULL,
    integration_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(audience_id, cabinet_id)
);

CREATE INDEX idx_audience_requests_audience_id ON audience_requests(audience_id);
CREATE INDEX idx_audience_integrations_audience_id ON audience_integrations(audience_id);