-- +goose Up
-- Audit database initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerations
CREATE TYPE audit_action AS ENUM (
    'user_register', 'user_login', 'user_logout',
    'wallet_create', 'wallet_archive',
    'transaction_create', 'transaction_confirm', 'transaction_fail',
    'exchange_quote', 'exchange_execute', 'exchange_complete',
    'kyc_submit', 'kyc_document_upload', 'kyc_approve', 'kyc_reject',
    'profile_update', 'password_change', '2fa_enable', '2fa_disable',
    'withdrawal_whitelist_add', 'withdrawal_whitelist_remove'
);

CREATE TYPE security_event AS ENUM (
    'login_success', 'login_failure', 'login_suspicious',
    'password_reset_request', 'password_reset_complete',
    '2fa_enabled', '2fa_disabled', '2fa_failure',
    'session_created', 'session_expired', 'session_revoked',
    'ip_blacklist_trigger', 'rate_limit_exceeded',
    'unusual_location', 'device_change',
    'withdrawal_whitelist_violation',
    'api_key_created', 'api_key_revoked'
);

CREATE TYPE security_severity AS ENUM ('info', 'warning', 'high', 'critical');

-- Audit logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    action audit_action NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB NOT NULL DEFAULT '{}'::JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    result VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);

-- Security logs
CREATE TABLE security_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    event_type security_event NOT NULL,
    severity security_severity NOT NULL,
    ip_address INET,
    user_agent TEXT,
    location_country VARCHAR(2),
    location_city VARCHAR(100),
    details JSONB NOT NULL DEFAULT '{}'::JSONB,
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_security_logs_user_id ON security_logs(user_id);
CREATE INDEX idx_security_logs_event_type ON security_logs(event_type);
CREATE INDEX idx_security_logs_severity ON security_logs(severity);
CREATE INDEX idx_security_logs_timestamp ON security_logs(timestamp DESC);
CREATE INDEX idx_security_logs_ip_address ON security_logs(ip_address);
CREATE INDEX idx_security_logs_unresolved ON security_logs(is_resolved) WHERE is_resolved = FALSE;

-- API audit
CREATE TABLE api_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    request_body_hash VARCHAR(64),
    response_body_hash VARCHAR(64),
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    duration_ms INTEGER NOT NULL,
    ip_address INET,
    user_agent TEXT,
    api_key_id UUID,
    rate_limit_remaining INTEGER,
    error_code VARCHAR(50),
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_audit_user_id ON api_audit(user_id);
CREATE INDEX idx_api_audit_endpoint ON api_audit(endpoint);
CREATE INDEX idx_api_audit_status_code ON api_audit(status_code);
CREATE INDEX idx_api_audit_timestamp ON api_audit(timestamp DESC);
CREATE INDEX idx_api_audit_ip_address ON api_audit(ip_address);
CREATE INDEX idx_api_audit_errors ON api_audit(error_code) WHERE error_code IS NOT NULL;
