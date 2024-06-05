-- Store only the accepted connections
CREATE TABLE connections (
    id SERIAL PRIMARY KEY,
    sender_id VARCHAR NOT NULL ,
    recipient_id VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    UNIQUE (sender_id, recipient_id)
);
