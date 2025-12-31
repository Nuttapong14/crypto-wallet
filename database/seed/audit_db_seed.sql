-- Seed data for audit_db delivering representative audit trails.

BEGIN;

INSERT INTO audit_logs (
    id, user_id, action, resource_type, resource_id,
    details, ip_address, user_agent, session_id,
    result, error_message, timestamp
) VALUES
    ('70000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
        'user_login', 'session', 'aaaaaaaa-0000-0000-0000-aaaaaaaaaaaa',
        '{"method":"password"}', '192.0.2.10', 'Mozilla/5.0 (Macintosh)', 'sess-alice-001',
        'success', NULL, '2025-01-08T08:55:00Z'),
    ('70000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
        'wallet_create', 'wallet', '20000000-0000-0000-0000-000000000002',
        '{"chain":"ETH"}', '192.0.2.10', 'Mozilla/5.0 (Macintosh)', 'sess-alice-001',
        'success', NULL, '2025-01-08T09:10:00Z'),
    ('70000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002',
        'transaction_create', 'transaction', '30000000-0000-0000-0000-000000000003',
        '{"amount":"25","chain":"SOL"}', '198.51.100.15', 'Mozilla/5.0 (Windows NT 10.0)', 'sess-bob-001',
        'success', NULL, '2025-01-09T08:05:10Z')
ON CONFLICT (id) DO NOTHING;

INSERT INTO security_logs (
    id, user_id, event_type, severity, ip_address, user_agent,
    location_country, location_city, details, is_resolved,
    resolved_at, resolved_by, timestamp
) VALUES
    ('71000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
        '2fa_enabled', 'info', '192.0.2.10', 'Mozilla/5.0 (Macintosh)',
        'US', 'New York', '{"issuer":"Atlas Wallet"}', TRUE,
        '2025-01-06T12:05:00Z', NULL, '2025-01-06T12:00:00Z'),
    ('71000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002',
        'login_failure', 'warning', '203.0.113.44', 'Mozilla/5.0 (Android 13)',
        'GB', 'London', '{"attempts":2}', FALSE,
        NULL, NULL, '2025-01-09T07:45:00Z')
ON CONFLICT (id) DO NOTHING;

INSERT INTO api_audit (
    id, user_id, endpoint, method, status_code,
    request_body_hash, response_body_hash, request_size_bytes,
    response_size_bytes, duration_ms, ip_address, user_agent,
    api_key_id, rate_limit_remaining, error_code, error_message, timestamp
) VALUES
    ('72000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
        '/api/v1/wallets', 'POST', 201,
        'hash-create-wallet', 'hash-wallet-response', 512, 1024, 120,
        '192.0.2.10', 'AtlasWallet/1.0', NULL, 99, NULL, NULL, '2025-01-08T09:10:02Z'),
    ('72000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002',
        '/api/v1/transactions', 'POST', 422,
        'hash-send-invalid', 'hash-error-response', 768, 256, 80,
        '198.51.100.15', 'AtlasWallet/1.0', NULL, 98, 'VALIDATION_ERROR', 'amount is required', '2025-01-09T08:03:00Z')
ON CONFLICT (id) DO NOTHING;

COMMIT;
