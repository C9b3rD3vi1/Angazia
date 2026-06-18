-- migrations/014_create_testimonials.sql

CREATE TABLE IF NOT EXISTS testimonials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name VARCHAR(255) NOT NULL,
    user_title VARCHAR(255),
    company_name VARCHAR(255),
    content TEXT NOT NULL,
    rating INT DEFAULT 0,
    is_approved BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    role VARCHAR(20) DEFAULT 'employee',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_testimonials_user_id (user_id),
    INDEX idx_testimonials_is_approved (is_approved),
    INDEX idx_testimonials_role (role),
    INDEX idx_testimonials_is_featured (is_featured)
);
