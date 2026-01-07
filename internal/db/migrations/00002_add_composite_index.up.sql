-- Create composite index for optimized read queries
-- This index covers the WHERE clause: short_code = ? AND expires_at > ?
-- Allows PostgreSQL to satisfy the query entirely from the index without reading the row
CREATE INDEX idx_url_records_short_code_expires_at ON url_records(short_code, expires_at);

-- Add comment explaining the index purpose
COMMENT ON INDEX idx_url_records_short_code_expires_at IS 'Composite index for efficient lookups by short_code with expiration filtering';
