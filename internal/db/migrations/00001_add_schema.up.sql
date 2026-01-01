-- Create url_records table
CREATE TABLE url_records (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE UNIQUE INDEX idx_url_records_short_code ON url_records(short_code);
CREATE INDEX idx_url_records_expires_at ON url_records(expires_at);

-- Add comment to table
COMMENT ON TABLE url_records IS 'Stores shortened URL mappings with expiration and soft delete support';
