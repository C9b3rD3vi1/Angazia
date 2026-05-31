-- migrations/011_create_admin_tables.sql

-- Admin action logs
CREATE TABLE IF NOT EXISTS admin_action_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_admin_action_logs_admin_id (admin_id),
    INDEX idx_admin_action_logs_action (action),
    INDEX idx_admin_action_logs_entity_type (entity_type),
    INDEX idx_admin_action_logs_created_at (created_at)
);

-- Moderation queue
CREATE TABLE IF NOT EXISTS moderation_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    reason TEXT,
    submitted_by UUID REFERENCES users(id),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_moderation_queue_entity_type (entity_type),
    INDEX idx_moderation_queue_status (status),
    INDEX idx_moderation_queue_entity_id (entity_id)
);

-- System settings
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    type VARCHAR(50),
    description TEXT,
    category VARCHAR(100),
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_system_settings_key (key),
    INDEX idx_system_settings_category (category)
);

-- Report reasons
CREATE TABLE IF NOT EXISTS report_reasons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    entity_type VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_report_reasons_entity_type (entity_type),
    INDEX idx_report_reasons_is_active (is_active)
);

-- Insert default report reasons
INSERT INTO report_reasons (name, description, entity_type, sort_order) VALUES
    ('Spam', 'Irrelevant or promotional content', 'job', 1),
    ('Inappropriate', 'Offensive or harmful content', 'job', 2),
    ('Fake Job', 'Job posting is fraudulent or misleading', 'job', 3),
    ('Expired Job', 'Job is no longer available', 'job', 4),
    ('Fake Profile', 'Suspicious or fake user profile', 'candidate', 1),
    ('Inappropriate Content', 'Offensive profile content', 'candidate', 2),
    ('Fake Company', 'Company information is fraudulent', 'company', 1),
    ('Scam', 'Suspected scam activity', 'company', 2);

-- Search queries table
CREATE TABLE IF NOT EXISTS search_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query TEXT,
    filters JSONB,
    entity_type VARCHAR(50) DEFAULT 'job',
    results_count INT DEFAULT 0,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_search_queries_user_id (user_id),
    INDEX idx_search_queries_entity_type (entity_type),
    INDEX idx_search_queries_created_at (created_at)
);

-- Saved searches table
CREATE TABLE IF NOT EXISTS saved_searches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    filters JSONB NOT NULL,
    entity_type VARCHAR(50) DEFAULT 'job',
    frequency VARCHAR(50) DEFAULT 'daily',
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_saved_searches_user_id (user_id),
    INDEX idx_saved_searches_entity_type (entity_type),
    INDEX idx_saved_searches_is_active (is_active)
);

-- Create full-text search indexes
CREATE INDEX IF NOT EXISTS idx_jobs_search_vector ON jobs USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));
CREATE INDEX IF NOT EXISTS idx_employee_search_vector ON employee_profiles USING GIN(to_tsvector('english', COALESCE(full_name, '') || ' ' || COALESCE(headline, '') || ' ' || COALESCE(bio, '')));
CREATE INDEX IF NOT EXISTS idx_employer_search_vector ON employer_profiles USING GIN(to_tsvector('english', COALESCE(company_name, '') || ' ' || COALESCE(company_description, '')));

-- Create triggers
CREATE TRIGGER update_moderation_queue_updated_at 
    BEFORE UPDATE ON moderation_queue 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_settings_updated_at 
    BEFORE UPDATE ON system_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_saved_searches_updated_at 
    BEFORE UPDATE ON saved_searches 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();