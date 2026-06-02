-- migrations/013_create_twofa_tables.sql

-- 2FA secrets table
CREATE TABLE IF NOT EXISTS two_fa_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    secret TEXT NOT NULL,
    method VARCHAR(20) DEFAULT 'app',
    phone_number VARCHAR(20),
    email VARCHAR(255),
    backup_codes JSONB,
    is_enabled BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    recovery_codes_used INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_two_fa_secrets_user_id (user_id),
    INDEX idx_two_fa_secrets_is_enabled (is_enabled)
);

-- Trusted devices table (optional, for "remember this device")
CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    user_agent TEXT,
    ip_address VARCHAR(45),
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, device_id),
    INDEX idx_trusted_devices_user_id (user_id),
    INDEX idx_trusted_devices_device_id (device_id)
);