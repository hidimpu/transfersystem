-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_account_id) REFERENCES accounts(account_id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(account_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_transactions_source ON transactions(source_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_destination ON transactions(destination_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at); 