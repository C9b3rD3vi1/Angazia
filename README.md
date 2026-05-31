# Angazia - Kenyan Tech Talent Marketplace

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-2.52-00ADD8?style=for-the-badge&logo=fiber)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![OpenAI](https://img.shields.io/badge/OpenAI-Enabled-412991?style=for-the-badge&logo=openai)](https://openai.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

> **Connecting Kenyan tech talent with great opportunities through AI-powered matching**

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Database Schema](#database-schema)
- [Deployment](#deployment)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

Angazia is a comprehensive job marketplace designed specifically for the Kenyan tech industry. It bridges the gap between talented developers and employers by leveraging AI-powered matching, GitHub integration, and intelligent analytics.

### The Problem We Solve

- **For Job Seekers**: Generic applications get lost, no way to showcase real skills, endless manual searching
- **For Employers**: Flood of unqualified applicants, no way to verify skills, time-consuming screening process

### Our Solution

- **AI-Powered Matching**: Intelligent job-candidate matching using OpenAI
- **GitHub Integration**: Showcase real code and contributions
- **Smart Analytics**: Data-driven insights for both candidates and employers
- **Automated Workflows**: From application to hiring, all in one platform

---

## ✨ Features

### 🏗️ Core Platform

| Feature | Status | Description |
|---------|--------|-------------|
| User Authentication | ✅ | JWT-based auth with refresh tokens, email verification |
| Role Management | ✅ | Admin, Employer, Employee roles with permissions |
| Profile Management | ✅ | Complete profiles with skills, experience, portfolio |
| GitHub Integration | ✅ | OAuth, repo sync, contribution tracking, activity scoring |

### 💼 Job Management

| Feature | Status | Description |
|---------|--------|-------------|
| Job Posting | ✅ | Create, edit, delete job listings |
| Job Search | ✅ | Full-text search with advanced filters |
| Save Jobs | ✅ | Candidates can save interesting jobs |
| Job Analytics | ✅ | View counts, application rates, conversion metrics |

### 📝 Application System

| Feature | Status | Description |
|---------|--------|-------------|
| Apply to Jobs | ✅ | Submit applications with cover letters |
| Application Tracking | ✅ | Real-time status updates |
| Shortlist/Reject | ✅ | Employer decision management |
| Interview Scheduling | ✅ | Calendar integration, email notifications |
| Bulk Actions | ✅ | Mass shortlist/reject applications |

### 🤖 AI & Matching

| Feature | Status | Description |
|---------|--------|-------------|
| Smart Matching | ✅ | AI-powered job-candidate matching scores |
| Cover Letter Generation | ✅ | Personalized AI-generated cover letters |
| Skills Gap Analysis | ✅ | Identify missing skills with learning resources |
| Interview Questions | ✅ | Role-specific question generation |
| Pluggable AI | ✅ | Support for OpenAI, Anthropic, Gemini, Local LLMs |

### 📄 Resume Parsing

| Feature | Status | Description |
|---------|--------|-------------|
| PDF/DOCX Parsing | ✅ | Extract text from uploaded resumes |
| Skill Extraction | ✅ | Automatic identification of technical skills |
| Experience Parsing | ✅ | Work history, years of experience |
| Contact Extraction | ✅ | Email, phone, LinkedIn, GitHub URLs |
| Profile Wizard | ✅ | Step-by-step profile completion guide |

### 🏢 Employer Features

| Feature | Status | Description |
|---------|--------|-------------|
| Company Profile | ✅ | Branding, description, social links |
| Verification | ✅ | Document submission, trust badges |
| Team Management | ✅ | Invite members, role-based access |
| Company Reviews | ✅ | Candidate reviews, ratings, helpful votes |
| Talent Pool | ✅ | Save candidates, tags, notes, status tracking |

### 📊 Analytics

| Feature | Status | Description |
|---------|--------|-------------|
| Application Trends | ✅ | Daily/weekly/monthly volume tracking |
| Conversion Funnel | ✅ | View → Apply → Shortlist → Interview → Hire |
| Job Performance | ✅ | Per-job metrics and comparisons |
| Time to Hire | ✅ | Average/median days to hire |
| Candidate Analytics | ✅ | Profile strength, success rates, market positioning |
| Skill Gap Analysis | ✅ | In-demand skills, learning recommendations |
| Export Reports | ✅ | CSV/JSON data export |

### 🔔 Notifications

| Feature | Status | Description |
|---------|--------|-------------|
| Email Notifications | ✅ | All application-related emails with HTML templates |
| Job Alerts | ✅ | Daily/weekly/instant job match digests |
| In-App Notifications | ⏳ | Coming soon |
| Real-time WebSocket | ⏳ | Coming soon |

---

## 🛠️ Tech Stack

### Backend
Go 1.21+ - Primary language

Fiber v2.52 - Web framework

GORM - ORM library

PostgreSQL 15 - Primary database

Redis 7.0 - Caching & token blacklisting

JWT - Authentication

WebSocket - Real-time features (coming soon)

text

### AI & Processing
OpenAI GPT-4 - Primary AI provider

Anthropic Claude - Alternative AI provider

Google Gemini - Alternative AI provider

Local LLM - Self-hosted option

Colly - Web scraping (deprecated)

text

### Email Providers
SendGrid - Primary email provider

Resend - Alternative email provider

SMTP - Fallback email provider

text

### Deployment
Docker - Containerization

Docker Compose - Multi-container orchestration

Nginx - Reverse proxy & SSL

GitHub Actions - CI/CD pipeline

text

---

## 🏗️ Architecture

### High-Level Architecture
┌─────────────────────────────────────────────────────────────┐
│ Client Layer │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│ │ Web App │ │ Mobile │ │ API │ │ Admin │ │
│ │ │ │ │ │ Clients │ │ Portal │ │
│ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ │
└───────┼─────────────┼─────────────┼─────────────┼─────────┘
│ │ │ │
▼ ▼ ▼ ▼
┌─────────────────────────────────────────────────────────────┐
│ API Gateway (Nginx) │
│ Rate Limiting / SSL Termination / CORS │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Application Layer (Fiber) │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ │
│ │ Handlers │ │ Services │ │ Middleware │ │
│ │ (HTTP) │─▶│ (Business) │ │ (Auth/Rate) │ │
│ └──────────────┘ └──────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────┐
│ Repository Layer (GORM) │
│ Database Operations / Query Building │
└─────────────────────────────────────────────────────────────┘
│
├──────────────┬──────────────┬──────────────┐
▼ ▼ ▼ ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ PostgreSQL │ │ Redis │ │ S3 │ │ OpenAI │
│ (Primary) │ │ (Cache) │ │ (Storage) │ │ (AI) │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘

text

### Directory Structure
angazia/
├── cmd/
│ ├── api/ # API server entry point
│ └── worker/ # Background worker (cron jobs)
│
├── internal/
│ ├── config/ # Configuration management
│ ├── models/ # Domain models
│ ├── repository/ # Database operations
│ │ ├── user_repo.go
│ │ ├── job_repo.go
│ │ ├── application_repo.go
│ │ ├── github_repo.go
│ │ ├── match_repo.go
│ │ ├── alert_repo.go
│ │ ├── company_repo.go
│ │ ├── talent_pool_repo.go
│ │ └── analytics_repo.go
│ │
│ ├── services/ # Business logic
│ │ ├── auth_service.go
│ │ ├── job_service.go
│ │ ├── application_service.go
│ │ ├── github_service.go
│ │ ├── matching_service.go
│ │ ├── alert_service.go
│ │ ├── company_service.go
│ │ ├── talent_pool_service.go
│ │ ├── analytics_service.go
│ │ └── email_service.go
│ │
│ ├── handlers/ # HTTP handlers
│ │ ├── auth_handler.go
│ │ ├── job_handler.go
│ │ ├── application_handler.go
│ │ ├── github_handler.go
│ │ ├── matching_handler.go
│ │ ├── alert_handler.go
│ │ ├── company_handler.go
│ │ ├── talent_pool_handler.go
│ │ └── analytics_handler.go
│ │
│ ├── routes/ # Route definitions
│ ├── middleware/ # HTTP middleware
│ └── pkg/ # Shared packages
│ ├── ai/ # AI providers
│ ├── database/ # DB connection
│ ├── github/ # GitHub API client
│ ├── parser/ # Resume parser
│ └── utils/ # Utilities
│
├── web/
│ ├── templates/ # HTML templates
│ │ └── emails/ # Email templates
│ └── static/ # Static assets
│
├── migrations/ # SQL migrations
├── scripts/ # Utility scripts
├── docker/ # Docker configuration
├── docs/ # Documentation
└── tests/ # Test files

text

---

## 📦 Installation

### Prerequisites

```bash
# Required
Go 1.21 or higher
PostgreSQL 15 or higher
Redis 7.0 or higher
Make

# Optional (for AI features)
OpenAI API Key
Anthropic API Key
Google Gemini API Key

# Optional (for email)
SendGrid API Key
Resend API Key
SMTP credentials
Quick Start
bash
# 1. Clone the repository
git clone https://github.com/C9b3rD3vi1/angazia.git
cd angazia

# 2. Copy environment configuration
cp .env.example .env

# 3. Edit .env with your configuration
vim .env

# 4. Create database
createdb angazia_db

# 5. Install dependencies
go mod download
go mod tidy

# 6. Run migrations
go run cmd/migrate/main.go up

# 7. Seed development data (optional)
go run scripts/seed.go

# 8. Start the API server
go run cmd/api/main.go

# 9. Start the background worker (in another terminal)
go run cmd/worker/main.go
Docker Deployment
bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
⚙️ Configuration
Environment Variables
env
# ==================== SERVER CONFIGURATION ====================
PORT=3000
ENVIRONMENT=development          # development, staging, production
APP_NAME=Angazia
APP_VERSION=1.0.0
APP_URL=http://localhost:3000
APP_DOMAIN=localhost
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:8080

# ==================== DATABASE CONFIGURATION ====================
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=angazia_db
DB_SSL_MODE=disable

# ==================== REDIS CONFIGURATION ====================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# ==================== JWT CONFIGURATION ====================
JWT_SECRET=your-super-secret-key-change-this
JWT_EXPIRY_HOURS=24

# ==================== GITHUB OAUTH ====================
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:3000/api/v1/github/callback
GITHUB_WEBHOOK_SECRET=your_webhook_secret

# ==================== AI PROVIDER ====================
AI_PROVIDER=openai               # openai, anthropic, gemini, local
AI_MODEL=gpt-4-turbo-preview

# OpenAI
OPENAI_API_KEY=your_openai_api_key

# Anthropic (optional)
ANTHROPIC_API_KEY=your_anthropic_api_key

# Google Gemini (optional)
GEMINI_API_KEY=your_gemini_api_key

# Local LLM (optional)
LOCAL_LLM_URL=http://localhost:8080/v1/completions

# ==================== EMAIL PROVIDER ====================
EMAIL_PROVIDER=smtp              # sendgrid, resend, smtp

# SendGrid
SENDGRID_API_KEY=your_sendgrid_api_key

# Resend
RESEND_API_KEY=your_resend_api_key

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@angazia.com
SMTP_PASSWORD=your_smtp_password
SMTP_FROM_NAME=Angazia
SMTP_FROM_EMAIL=noreply@angazia.com

# ==================== SECURITY ====================
REQUIRE_EMAIL_VERIFICATION=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60

# ==================== FILE UPLOADS ====================
MAX_FILE_SIZE=10485760          # 10MB
ALLOWED_FILE_TYPES=.pdf,.doc,.docx,.jpg,.jpeg,.png

# ==================== M-PESA (Future) ====================
MPESA_CONSUMER_KEY=your_mpesa_consumer_key
MPESA_CONSUMER_SECRET=your_mpesa_consumer_secret
MPESA_SHORTCODE=174379
MPESA_PASSKEY=your_mpesa_passkey
MPESA_ENVIRONMENT=sandbox

# ==================== FEATURE FLAGS ====================
MAX_JOB_POSTS_FREE=3
MAX_JOB_POSTS_PRO=20
MAX_JOB_POSTS_ENTERPRISE=100
PAGE_SIZE=20
📚 API Documentation
Authentication Endpoints
http
POST   /api/v1/auth/register           # Register new user
POST   /api/v1/auth/login              # Login user
POST   /api/v1/auth/logout             # Logout user
POST   /api/v1/auth/refresh            # Refresh access token
POST   /api/v1/auth/forgot-password    # Request password reset
POST   /api/v1/auth/reset-password     # Reset password
GET    /api/v1/auth/verify-email/:token # Verify email address
POST   /api/v1/auth/change-password    # Change password (auth required)
POST   /api/v1/auth/resend-verification # Resend verification email
User Profile Endpoints
http
GET    /api/v1/profile                 # Get user profile
PUT    /api/v1/profile                 # Update user profile
GET    /api/v1/employee/dashboard      # Employee dashboard
GET    /api/v1/employer/dashboard      # Employer dashboard
Job Endpoints
http
# Public (no auth required)
GET    /api/v1/jobs                    # List jobs (paginated)
GET    /api/v1/jobs/featured           # Get featured jobs
GET    /api/v1/jobs/recent             # Get recent jobs
GET    /api/v1/jobs/search             # Search jobs
GET    /api/v1/jobs/:id                # Get job details
GET    /api/v1/jobs/:id/similar        # Get similar jobs

# Authenticated
POST   /api/v1/jobs/:id/save           # Save job (candidate)
DELETE /api/v1/jobs/:id/save           # Unsave job (candidate)
GET    /api/v1/employee/saved-jobs     # List saved jobs (candidate)

# Employer only
POST   /api/v1/employer/jobs           # Create job
GET    /api/v1/employer/jobs           # List employer jobs
PUT    /api/v1/employer/jobs/:id       # Update job
DELETE /api/v1/employer/jobs/:id       # Delete job
POST   /api/v1/employer/jobs/:id/close # Close job
Application Endpoints
http
# Candidate
POST   /api/v1/employee/applications   # Submit application
GET    /api/v1/employee/applications   # List my applications
GET    /api/v1/applications/:id        # Get application details
POST   /api/v1/applications/:id/withdraw # Withdraw application

# Employer
GET    /api/v1/employer/applications   # List all company applications
GET    /api/v1/employer/jobs/:jobId/applications # List job applications
POST   /api/v1/employer/applications/:id/shortlist # Shortlist
POST   /api/v1/employer/applications/:id/reject    # Reject
POST   /api/v1/employer/applications/:id/interview # Schedule interview
POST   /api/v1/employer/applications/bulk-shortlist # Bulk shortlist
GitHub Integration Endpoints
http
GET    /api/v1/github/auth             # Start GitHub OAuth
GET    /api/v1/github/callback         # GitHub OAuth callback
POST   /api/v1/github/webhook          # GitHub webhook receiver

# Authenticated
POST   /api/v1/github/connect          # Connect GitHub account
POST   /api/v1/github/disconnect       # Disconnect GitHub account
POST   /api/v1/github/sync             # Sync GitHub data
GET    /api/v1/github/profile          # Get GitHub profile
GET    /api/v1/github/repos            # List GitHub repositories
GET    /api/v1/github/contributions    # Get contribution calendar
AI Matching Endpoints
http
# Candidate
GET    /api/v1/employee/matches/jobs   # Get job recommendations
POST   /api/v1/employee/matches/cover-letter # Generate cover letter
GET    /api/v1/employee/matches/skills-gap/:jobId # Analyze skills gap

# Employer
GET    /api/v1/employer/matches/candidates/:jobId # Get candidate recommendations
GET    /api/v1/employer/matches/interview-questions/:jobId # Generate questions

# General
GET    /api/v1/matches/analysis/:jobId/:employeeId # Detailed match analysis
Job Alert Endpoints
http
POST   /api/v1/alerts/search           # Create saved search
GET    /api/v1/alerts                  # List saved searches
GET    /api/v1/alerts/:id              # Get saved search
PUT    /api/v1/alerts/:id              # Update saved search
DELETE /api/v1/alerts/:id              # Delete saved search
POST   /api/v1/alerts/:id/test         # Send test alert
GET    /api/v1/alerts/settings         # Get alert settings
PUT    /api/v1/alerts/settings         # Update alert settings
GET    /api/v1/alerts/history          # Get alert history
Company Management Endpoints
http
# Public
GET    /api/v1/companies/:id           # Public company profile
GET    /api/v1/companies/:id/badges    # Get company badges
GET    /api/v1/companies/:id/reviews   # Get company reviews
GET    /api/v1/companies/:id/reviews/stats # Get review stats

# Employer only
GET    /api/v1/employer/company        # Get company profile
PUT    /api/v1/employer/company        # Update company profile
POST   /api/v1/employer/company/logo   # Upload company logo
POST   /api/v1/employer/company/verify # Submit verification
GET    /api/v1/employer/company/verification # Get verification status

# Team management
GET    /api/v1/employer/team           # List team members
POST   /api/v1/employer/team/invite    # Invite team member
DELETE /api/v1/employer/team/:memberId # Remove team member

# Reviews (authenticated)
POST   /api/v1/companies/:id/reviews   # Submit review
POST   /api/v1/reviews/:id/helpful     # Mark review helpful
Talent Pool Endpoints
http
POST   /api/v1/employer/talent-pools   # Create talent pool
GET    /api/v1/employer/talent-pools   # List talent pools
GET    /api/v1/employer/talent-pools/:id # Get talent pool
PUT    /api/v1/employer/talent-pools/:id # Update talent pool
DELETE /api/v1/employer/talent-pools/:id # Delete talent pool
GET    /api/v1/employer/talent-pools/:id/stats # Get pool stats

# Candidate management
GET    /api/v1/employer/talent-pools/:id/candidates # List candidates
POST   /api/v1/employer/talent-pools/:id/candidates # Add candidate
PUT    /api/v1/employer/talent-pools/:poolId/candidates/:candidateId # Update candidate
DELETE /api/v1/employer/talent-pools/:poolId/candidates/:candidateId # Remove candidate
POST   /api/v1/employer/talent-pools/:poolId/candidates/:candidateId/contact # Mark contacted
POST   /api/v1/employer/talent-pools/:poolId/candidates/:candidateId/hire # Mark hired
Analytics Endpoints
http
# Employer analytics
GET    /api/v1/employer/analytics/dashboard # Complete dashboard
GET    /api/v1/employer/analytics/trends    # Application trends
GET    /api/v1/employer/analytics/funnel    # Conversion funnel
GET    /api/v1/employer/analytics/jobs      # Job performance
GET    /api/v1/employer/analytics/time-to-hire # Time to hire
GET    /api/v1/employer/analytics/quality   # Application quality
GET    /api/v1/employer/analytics/sources   # Source analytics
GET    /api/v1/employer/analytics/export    # Export data

# Candidate analytics
GET    /api/v1/employee/analytics/dashboard # Complete dashboard
GET    /api/v1/employee/analytics/profile-strength # Profile strength
GET    /api/v1/employee/analytics/applications/stats # Application stats
GET    /api/v1/employee/analytics/success-rates # Success rates
GET    /api/v1/employee/analytics/skill-gap # Skill gap analysis
GET    /api/v1/employee/analytics/market-positioning # Market positioning
GET    /api/v1/employee/analytics/recommendations # Recommendations
Resume Parsing Endpoints
http
POST   /api/v1/employee/resume/upload   # Upload and parse resume
GET    /api/v1/employee/profile/completion # Get profile completion
GET    /api/v1/employee/skills/suggested # Get suggested skills
GET    /api/v1/employee/profile/wizard  # Get profile wizard
🗄️ Database Schema
Core Tables
Table	Description
users	User accounts (employees, employers, admins)
employee_profiles	Job seeker profiles
employer_profiles	Company profiles
jobs	Job postings
applications	Job applications
saved_jobs	Jobs saved by candidates
GitHub Integration Tables
Table	Description
github_profiles	GitHub user profiles
github_repositories	GitHub repositories
github_contributions	Daily contribution counts
github_tokens	Encrypted OAuth tokens
AI & Matching Tables
Table	Description
matches	Job-candidate match records
match_feedback	User feedback on matches
match_settings	User matching preferences
Alerts & Notifications
Table	Description
saved_searches	Saved job search criteria
alert_history	Sent alert records
alert_settings	User notification preferences
unsubscribe_tokens	Email unsubscribe tokens
Company & Reviews
Table	Description
company_verifications	Verification documents and status
trust_badges	Company trust badges
company_reviews	Company ratings and reviews
team_invitations	Team member invitations
Talent Pool
Table	Description
talent_pools	Candidate collections
talent_pool_candidates	Candidates in pools
Analytics
Table	Description
company_analytics	Daily company metrics
job_views	Job view tracking
search_history	User search queries
🚀 Deployment
Production Deployment
bash
# 1. Set production environment
export ENVIRONMENT=production

# 2. Build the application
go build -o angazia-api cmd/api/main.go
go build -o angazia-worker cmd/worker/main.go

# 3. Run database migrations
./angazia-api migrate up

# 4. Start services
./angazia-api &
./angazia-worker &

# Or use systemd (recommended)
sudo systemctl start angazia-api
sudo systemctl start angazia-worker
Systemd Service Files
/etc/systemd/system/angazia-api.service

ini
[Unit]
Description=Angazia API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=angazia
WorkingDirectory=/opt/angazia
ExecStart=/opt/angazia/angazia-api
Restart=always
RestartSec=10
Environment="ENVIRONMENT=production"

[Install]
WantedBy=multi-user.target
/etc/systemd/system/angazia-worker.service

ini
[Unit]
Description=Angazia Background Worker
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=angazia
WorkingDirectory=/opt/angazia
ExecStart=/opt/angazia/angazia-worker
Restart=always
RestartSec=10
Environment="ENVIRONMENT=production"

[Install]
WantedBy=multi-user.target
Nginx Configuration
nginx
server {
    listen 80;
    server_name angazia.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name angazia.com;

    ssl_certificate /etc/letsencrypt/live/angazia.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/angazia.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /static {
        alias /opt/angazia/web/static;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location /uploads {
        alias /opt/angazia/uploads;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
Docker Compose Production
yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: angazia
      POSTGRES_USER: angazia
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - angazia

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - angazia

  api:
    build:
      context: .
      dockerfile: Dockerfile.api
    environment:
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis
    networks:
      - angazia
    restart: always

  worker:
    build:
      context: .
      dockerfile: Dockerfile.worker
    environment:
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis
    networks:
      - angazia
    restart: always

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
      - ./ssl:/etc/nginx/ssl
      - ./web/static:/var/www/static
    depends_on:
      - api
    networks:
      - angazia
    restart: always

volumes:
  postgres_data:
  redis_data:

networks:
  angazia:
    driver: bridge
🧪 Testing
bash
# Run all tests
go test -v ./...

# Run unit tests
go test -v ./internal/...

# Run integration tests
go test -v ./tests/integration/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...
🤝 Contributing
Fork the repository

Create your feature branch (git checkout -b feature/amazing-feature)

Commit your changes (git commit -m 'Add amazing feature')

Push to the branch (git push origin feature/amazing-feature)

Open a Pull Request

Code Style
Follow Go standard formatting (gofmt)

Use descriptive variable names

Add comments for complex logic

Write tests for new features

📄 License
This project is licensed under the MIT License - see the LICENSE file for details.

🙏 Acknowledgments
Fiber - Fast Go web framework

GORM - Amazing ORM library

PostgreSQL - Powerful relational database

Redis - In-memory data store

OpenAI - AI capabilities

GitHub - OAuth and API integration

All Contributors - For making this project possible

📞 Support
Documentation: Wiki

Issues: GitHub Issues

Discussions: GitHub Discussions

Email: support@angazia.com

🚦 Status
https://img.shields.io/github/actions/workflow/status/C9b3rD3vi1/angazia/ci.yml?branch=main
https://img.shields.io/codecov/c/github/C9b3rD3vi1/angazia
https://goreportcard.com/badge/github.com/C9b3rD3vi1/angazia

Built with ❤️ for the Kenyan tech community

Empowering talent, connecting opportunities.

text

This README provides:

1. **Complete Project Overview** - What the platform does and why
2. **Detailed Feature List** - All implemented features with status
3. **Tech Stack Documentation** - All technologies used
4. **Architecture Diagrams** - System architecture and directory structure
5. **Installation Guide** - Step-by-step setup instructions
6. **Full API Documentation** - All endpoints with descriptions
7. **Database Schema** - All tables and their purposes
8. **Deployment Guide** - Production deployment instructions
9. **Testing Guide** - How to run tests
10. **Contributing Guidelines** - How to contribute

The README is production-ready and comprehensive! 🚀
