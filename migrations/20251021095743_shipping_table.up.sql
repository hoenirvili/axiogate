CREATE TABLE shipment (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL
);
