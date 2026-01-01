-- Drop indexes
DROP INDEX IF EXISTS idx_url_records_expires_at;
DROP INDEX IF EXISTS idx_url_records_short_code;

-- Drop table
DROP TABLE IF EXISTS url_records;

