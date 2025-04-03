CREATE TABLE bills (
    id VARCHAR(36) PRIMARY KEY,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    total_amount DECIMAL(19,4) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP,
    CONSTRAINT valid_currency CHECK (currency IN ('USD', 'GEL'))
);


CREATE TABLE line_items (
    id VARCHAR(36) PRIMARY KEY,
    bill_id VARCHAR(36) NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(19,4) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bill_id) REFERENCES bills(id) ON DELETE CASCADE
);