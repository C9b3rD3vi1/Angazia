-- migrations/008_create_company_tables.sql

-- Company verifications table
CREATE TABLE IF NOT EXISTS company_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    business_registration_number VARCHAR(100),
    tax_id VARCHAR(100),
    documents JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    rejection_reason TEXT,
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMP WITH TIME ZONE,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(company_id),
    INDEX idx_company_verifications_company_id (company_id),
    INDEX idx_company_verifications_status (status)
);

-- Trust badges table
CREATE TABLE IF NOT EXISTS trust_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    badge_type VARCHAR(50) NOT NULL,
    badge_name VARCHAR(100),
    description TEXT,
    icon_url VARCHAR(512),
    awarded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    INDEX idx_trust_badges_company_id (company_id),
    INDEX idx_trust_badges_badge_type (badge_type)
);

-- Company reviews table
CREATE TABLE IF NOT EXISTS company_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    content TEXT NOT NULL,
    pros TEXT,
    cons TEXT,
    would_recommend BOOLEAN DEFAULT FALSE,
    employment_status VARCHAR(50),
    is_verified BOOLEAN DEFAULT FALSE,
    helpful_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_company_reviews_company_id (company_id),
    INDEX idx_company_reviews_reviewer_id (reviewer_id),
    INDEX idx_company_reviews_rating (rating),
    UNIQUE(company_id, reviewer_id)
);

-- Team invitations table
CREATE TABLE IF NOT EXISTS team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    invited_by UUID NOT NULL REFERENCES users(id),
    accepted_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_team_invitations_company_id (company_id),
    INDEX idx_team_invitations_email (email),
    INDEX idx_team_invitations_token (token),
    INDEX idx_team_invitations_status (status)
);

-- Company analytics table
CREATE TABLE IF NOT EXISTS company_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    date DATE NOT NULL,
    profile_views INT DEFAULT 0,
    job_views INT DEFAULT 0,
    applications_received INT DEFAULT 0,
    searches_found INT DEFAULT 0,
    profile_completion_rate INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(company_id, date),
    INDEX idx_company_analytics_company_id (company_id),
    INDEX idx_company_analytics_date (date)
);

-- Update employer_profiles table to add company_id for team members
ALTER TABLE employer_profiles ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES employer_profiles(user_id);

-- Create trigger for updated_at
CREATE TRIGGER update_company_verifications_updated_at 
    BEFORE UPDATE ON company_verifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_company_reviews_updated_at 
    BEFORE UPDATE ON company_reviews 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();