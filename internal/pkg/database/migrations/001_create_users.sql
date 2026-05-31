-- 001_create_users.sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    role VARCHAR(50) NOT NULL, -- 'employee' or 'employer'
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 002_create_employee_profiles.sql
CREATE TABLE employee_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    github_username VARCHAR(255) UNIQUE,
    github_connected BOOLEAN DEFAULT FALSE,
    github_data JSONB, -- Store last fetch from GitHub API
    bio TEXT,
    skills TEXT[], -- Array of skills
    experience_level VARCHAR(50), -- entry, mid, senior
    resume_url TEXT,
    location VARCHAR(255),
    is_visible BOOLEAN DEFAULT TRUE,
    last_github_sync TIMESTAMP
);

-- 003_create_employer_profiles.sql
CREATE TABLE employer_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    company_website TEXT,
    company_linkedin TEXT,
    verification_status VARCHAR(50) DEFAULT 'pending',
    company_size VARCHAR(50),
    industry VARCHAR(100),
    location VARCHAR(255)
);

-- 004_create_jobs.sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    employer_id UUID NOT NULL REFERENCES employer_profiles(user_id),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    required_skills TEXT[],
    nice_to_have_skills TEXT[],
    experience_level VARCHAR(50),
    salary_min INTEGER,
    salary_max INTEGER,
    location VARCHAR(255),
    is_remote BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    posted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    views_count INTEGER DEFAULT 0,
    applications_count INTEGER DEFAULT 0
);

-- 005_create_applications.sql
CREATE TABLE applications (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profiles(user_id),
    status VARCHAR(50) DEFAULT 'pending', -- pending, viewed, shortlisted, rejected, hired
    match_score INTEGER, -- 1-100 calculated by AI
    cover_letter TEXT,
    employer_notes TEXT,
    applied_at TIMESTAMP DEFAULT NOW(),
    viewed_at TIMESTAMP,
    UNIQUE(job_id, employee_id)
);

-- 006_create_github_sync_log.sql
CREATE TABLE github_sync_log (
    id UUID PRIMARY KEY,
    employee_id UUID REFERENCES employee_profiles(user_id),
    synced_at TIMESTAMP DEFAULT NOW(),
    commits_last_30d INTEGER,
    repos_count INTEGER,
    top_languages JSONB,
    contribution_streak INTEGER,
    error_message TEXT
);