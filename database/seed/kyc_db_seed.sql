-- Seed data for kyc_db

BEGIN;

-- KYC profiles
INSERT INTO kyc_profiles (
    id, user_id, verification_level, status,
    first_name_encrypted, last_name_encrypted, date_of_birth_encrypted,
    nationality_encrypted, document_number_encrypted, address_encrypted,
    submitted_at, reviewed_at, approved_at, daily_limit_usd, monthly_limit_usd,
    created_at, updated_at
) VALUES
    ('50000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
        'full', 'approved',
        'ENC:alice:first', 'ENC:alice:last', 'ENC:1988-05-20',
        'ENC:US', 'ENC:ALICE12345', 'ENC:123 Demo Street, New York, US',
        '2025-01-05T08:00:00Z', '2025-01-06T10:00:00Z', '2025-01-06T12:00:00Z',
        50000, 500000, '2025-01-05T08:00:00Z', '2025-01-06T12:00:00Z'),
    ('50000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002',
        'basic', 'pending',
        'ENC:bob:first', 'ENC:bob:last', 'ENC:1990-09-02',
        'ENC:GB', NULL, 'ENC:45 Sample Road, London, UK',
        '2025-01-07T09:00:00Z', NULL, NULL,
        5000, 50000, '2025-01-07T09:00:00Z', '2025-01-07T09:00:00Z')
ON CONFLICT (user_id) DO NOTHING;

-- KYC documents
INSERT INTO kyc_documents (
    id, kyc_profile_id, document_type, file_path_encrypted,
    file_name_encrypted, file_size_bytes, file_hash, mime_type,
    status, uploaded_at, reviewed_at, rejection_reason, metadata, created_at, updated_at
) VALUES
    ('51000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001',
        'passport', 'ENC:s3://kyc/alice/passport.enc', 'ENC:passport.pdf',
        524288, 'hash-alice-passport', 'application/pdf', 'approved',
        '2025-01-05T08:10:00Z', '2025-01-06T10:00:00Z', NULL,
        '{"country":"US"}', '2025-01-05T08:10:00Z', '2025-01-06T10:00:00Z'),
    ('51000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000002',
        'national_id', 'ENC:s3://kyc/bob/id.enc', 'ENC:national-id.png',
        262144, 'hash-bob-id', 'image/png', 'pending',
        '2025-01-07T09:05:00Z', NULL, NULL,
        '{"notes":"Awaiting verification"}', '2025-01-07T09:05:00Z', '2025-01-07T09:05:00Z')
ON CONFLICT (id) DO NOTHING;

-- User risk scores
INSERT INTO user_risk_scores (
    id, user_id, risk_score, risk_level, risk_factors, aml_hits,
    last_screening_at, next_review_at, created_at, updated_at
) VALUES
    ('52000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
        18, 'low', '["verified_identity","consistent_activity"]', '[]',
        '2025-01-06T12:00:00Z', '2025-07-06T12:00:00Z',
        '2025-01-05T08:00:00Z', '2025-01-06T12:00:00Z'),
    ('52000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002',
        42, 'medium', '["pending_kyc","elevated_volume"]', '["worldcheck-screening"]',
        '2025-01-07T09:10:00Z', '2025-04-07T09:10:00Z',
        '2025-01-07T09:00:00Z', '2025-01-07T09:10:00Z')
ON CONFLICT (user_id) DO NOTHING;

-- Alert rules ensure AML automation can be demonstrated
INSERT INTO alert_rules (id, rule_name, rule_type, threshold_value,
                         time_window_minutes, is_active, severity, description,
                         created_at, updated_at)
VALUES
    ('53000000-0000-0000-0000-000000000001', 'high_value_daily', 'transaction_amount', 10000,
        60 * 24, TRUE, 'high', 'Triggers when a user exceeds $10k in daily volume.',
        '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z'),
    ('53000000-0000-0000-0000-000000000002', 'velocity_check', 'velocity', 5,
        60, TRUE, 'warning', 'Detects more than five transactions within an hour.',
        '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z')
ON CONFLICT (rule_name) DO NOTHING;

COMMIT;
