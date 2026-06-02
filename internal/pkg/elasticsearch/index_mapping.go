package elasticsearch

const JobIndexMapping = `{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "custom_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "stop", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "title": { 
        "type": "text",
        "analyzer": "custom_analyzer",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { "type": "text", "analyzer": "custom_analyzer" },
      "requirements": { "type": "text", "analyzer": "custom_analyzer" },
      "required_skills": { "type": "keyword" },
      "nice_to_have_skills": { "type": "keyword" },
      "experience_level": { "type": "keyword" },
      "min_experience": { "type": "integer" },
      "max_experience": { "type": "integer" },
      "location": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "is_remote": { "type": "boolean" },
      "is_hybrid": { "type": "boolean" },
      "employment_type": { "type": "keyword" },
      "salary_min": { "type": "integer" },
      "salary_max": { "type": "integer" },
      "salary_currency": { "type": "keyword" },
      "company_id": { "type": "keyword" },
      "company_name": { "type": "text", "analyzer": "custom_analyzer" },
      "company_industry": { "type": "keyword" },
      "is_active": { "type": "boolean" },
      "is_featured": { "type": "boolean" },
      "views_count": { "type": "integer" },
      "applications_count": { "type": "integer" },
      "posted_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}`

const CandidateIndexMapping = `{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "custom_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "stop", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "user_id": { "type": "keyword" },
      "full_name": { "type": "text", "analyzer": "custom_analyzer" },
      "headline": { "type": "text", "analyzer": "custom_analyzer" },
      "bio": { "type": "text", "analyzer": "custom_analyzer" },
      "skills": { "type": "keyword" },
      "experience_level": { "type": "keyword" },
      "years_of_experience": { "type": "integer" },
      "location": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "is_remote": { "type": "boolean" },
      "is_hybrid": { "type": "boolean" },
      "github_username": { "type": "keyword" },
      "github_activity": {
        "properties": {
          "public_repos": { "type": "integer" },
          "total_commits": { "type": "integer" },
          "followers": { "type": "integer" },
          "top_languages": { "type": "keyword" },
          "activity_score": { "type": "integer" }
        }
      },
      "is_available": { "type": "boolean" },
      "created_at": { "type": "date" }
    }
  }
}`

const CompanyIndexMapping = `{
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "custom_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "stop", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "name": { "type": "text", "analyzer": "custom_analyzer", "fields": { "keyword": { "type": "keyword" } } },
      "description": { "type": "text", "analyzer": "custom_analyzer" },
      "industry": { "type": "keyword" },
      "size": { "type": "keyword" },
      "location": { "type": "text", "fields": { "keyword": { "type": "keyword" } } },
      "website": { "type": "keyword" },
      "is_verified": { "type": "boolean" },
      "rating": { "type": "float" },
      "total_jobs": { "type": "integer" },
      "created_at": { "type": "date" }
    }
  }
}`