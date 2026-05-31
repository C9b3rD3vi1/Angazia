-- migrations/006_create_unsubscribe_tables.sql

-- Unsubscribe tokens table
CREATE TABLE IF NOT EXISTS unsubscribe_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    unsubscribed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_unsubscribe_tokens_email (email),
    INDEX idx_unsubscribe_tokens_token (token),
    INDEX idx_unsubscribe_tokens_user_id (user_id),
    INDEX idx_unsubscribe_tokens_is_active (is_active),
    INDEX idx_unsubscribe_tokens_expires_at (expires_at)
);

-- Email preferences table
CREATE TABLE IF NOT EXISTS email_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    job_alerts BOOLEAN DEFAULT TRUE,
    application_updates BOOLEAN DEFAULT TRUE,
    marketing_emails BOOLEAN DEFAULT TRUE,
    security_alerts BOOLEAN DEFAULT TRUE,
    newsletter BOOLEAN DEFAULT FALSE,
    digest_frequency VARCHAR(20) DEFAULT 'daily',
    last_digest_sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_email_preferences_user_id (user_id),
    INDEX idx_email_preferences_job_alerts (job_alerts)
);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_unsubscribe_tokens_updated_at 
    BEFORE UPDATE ON unsubscribe_tokens 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_email_preferences_updated_at 
    BEFORE UPDATE ON email_preferences 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Clean expired tokens periodically (run as cron job)
-- DELETE FROM unsubscribe_tokens WHERE expires_at < NOW() AND is_active = false;