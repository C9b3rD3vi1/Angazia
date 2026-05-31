-- migrations/009_create_talent_pool_tables.sql

-- Talent pools table
CREATE TABLE IF NOT EXISTS talent_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employer_id UUID NOT NULL REFERENCES employer_profiles(user_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(50) DEFAULT '#667eea',
    icon VARCHAR(50) DEFAULT 'users',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_talent_pools_employer_id (employer_id),
    INDEX idx_talent_pools_is_active (is_active)
);

-- Talent pool candidates table
CREATE TABLE IF NOT EXISTS talent_pool_candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    talent_pool_id UUID NOT NULL REFERENCES talent_pools(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles(user_id) ON DELETE CASCADE,
    match_score INT DEFAULT 0,
    notes TEXT,
    tags TEXT[],
    status VARCHAR(50) DEFAULT 'active',
    contacted_at TIMESTAMP WITH TIME ZONE,
    hired_at TIMESTAMP WITH TIME ZONE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(talent_pool_id, employee_id),
    INDEX idx_talent_pool_candidates_talent_pool_id (talent_pool_id),
    INDEX idx_talent_pool_candidates_employee_id (employee_id),
    INDEX idx_talent_pool_candidates_status (status),
    INDEX idx_talent_pool_candidates_match_score (match_score)
);

-- Create triggers for updated_at
CREATE TRIGGER update_talent_pools_updated_at 
    BEFORE UPDATE ON talent_pools 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_talent_pool_candidates_updated_at 
    BEFORE UPDATE ON talent_pool_candidates 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();