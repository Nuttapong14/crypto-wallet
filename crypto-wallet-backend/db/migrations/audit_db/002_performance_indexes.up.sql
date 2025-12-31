-- +goose Up
-- Indexes optimised for audit log lookups.

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_timestamp
    ON audit_logs(user_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_security_logs_event_timestamp
    ON security_logs(event_type, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_api_audit_endpoint_status
    ON api_audit(endpoint, status_code, timestamp DESC);
