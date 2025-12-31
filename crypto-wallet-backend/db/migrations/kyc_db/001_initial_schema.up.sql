-- +goose Up
-- KYC database initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerations
CREATE TYPE verification_level AS ENUM ('unverified', 'basic', 'full');
CREATE TYPE kyc_status AS ENUM ('not_started', 'pending', 'under_review', 'approved', 'rejected', 'expired');
CREATE TYPE document_type AS ENUM ('passport', 'national_id', 'drivers_license', 'proof_of_address', 'selfie');
CREATE TYPE document_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE rule_type AS ENUM ('transaction_amount', 'transaction_frequency', 'velocity', 'geography', 'pattern');

-- KYC profiles
CREATE TABLE kyc_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    verification_level verification_level NOT NULL DEFAULT 'unverified',
    status kyc_status NOT NULL DEFAULT 'not_started',
    first_name_encrypted TEXT,
    last_name_encrypted TEXT,
    date_of_birth_encrypted TEXT,
    nationality_encrypted TEXT,
    document_number_encrypted TEXT,
    address_encrypted TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    reviewer_notes TEXT,
    daily_limit_usd DECIMAL(15, 2) NOT NULL DEFAULT 500,
    monthly_limit_usd DECIMAL(15, 2) NOT NULL DEFAULT 5000,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_profiles_user_id ON kyc_profiles(user_id);
CREATE INDEX idx_kyc_profiles_status ON kyc_profiles(status);
CREATE INDEX idx_kyc_profiles_verification_level ON kyc_profiles(verification_level);
CREATE INDEX idx_kyc_profiles_submitted_at ON kyc_profiles(submitted_at);

-- KYC documents
CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_profile_id UUID NOT NULL REFERENCES kyc_profiles(id) ON DELETE CASCADE,
    document_type document_type NOT NULL,
    file_path_encrypted TEXT NOT NULL,
    file_name_encrypted TEXT NOT NULL,
    file_size_bytes INTEGER NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status document_status NOT NULL DEFAULT 'pending',
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_documents_kyc_profile_id ON kyc_documents(kyc_profile_id);
CREATE INDEX idx_kyc_documents_status ON kyc_documents(status);
CREATE INDEX idx_kyc_documents_document_type ON kyc_documents(document_type);
CREATE INDEX idx_kyc_documents_uploaded_at ON kyc_documents(uploaded_at);

-- User risk scores
CREATE TABLE user_risk_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    risk_score INTEGER NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    risk_level risk_level NOT NULL,
    risk_factors JSONB NOT NULL DEFAULT '[]'::JSONB,
    aml_hits JSONB NOT NULL DEFAULT '[]'::JSONB,
    last_screening_at TIMESTAMP WITH TIME ZONE,
    next_review_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_risk_scores_user_id ON user_risk_scores(user_id);
CREATE INDEX idx_user_risk_scores_risk_level ON user_risk_scores(risk_level);
CREATE INDEX idx_user_risk_scores_risk_score ON user_risk_scores(risk_score DESC);
CREATE INDEX idx_user_risk_scores_next_review_at ON user_risk_scores(next_review_at);

-- Alert rules
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name VARCHAR(200) NOT NULL UNIQUE,
    rule_type rule_type NOT NULL,
    threshold_value DECIMAL(15, 2) NOT NULL,
    time_window_minutes INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_is_active ON alert_rules(is_active);
CREATE INDEX idx_alert_rules_rule_type ON alert_rules(rule_type);
