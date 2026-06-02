package elasticsearch

import (
	"context"
	"encoding/json"
)

// CandidateSearchFilters are defined inline to avoid import cycles.

type CandidateSearchFilters struct {
	Keyword             string
	Skills              []string
	Location            string
	ExperienceLevel     string
	MinYearsOfExperience int
	MaxYearsOfExperience int
	IsRemote            *bool
	IsAvailable         *bool
	SortBy              string // "relevance", "experience", "created_at"
	Page                int
	Limit               int
}

type CompanySearchFilters struct {
	Keyword      string
	Industry     string
	Location     string
	Size         string
	MinRating    float64
	IsVerified   *bool
	SortBy       string // "relevance", "rating", "total_jobs"
	Page         int
	Limit        int
}

func BuildCandidateSearchQuery(filters CandidateSearchFilters) map[string]interface{} {
	from := (filters.Page - 1) * filters.Limit
	if from < 0 {
		from = 0
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	query := map[string]interface{}{
		"from": from,
		"size": filters.Limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []interface{}{},
				"filter": []interface{}{},
			},
		},
		"aggs": BuildCandidateAggregations(),
	}

	switch filters.SortBy {
	case "experience":
		query["sort"] = []interface{}{
			map[string]interface{}{"years_of_experience": "desc"},
		}
	case "created_at":
		query["sort"] = []interface{}{
			map[string]interface{}{"created_at": "desc"},
		}
	default:
		query["sort"] = []interface{}{
			map[string]interface{}{"_score": "desc"},
			map[string]interface{}{"created_at": "desc"},
		}
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

	if filters.Keyword != "" {
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     filters.Keyword,
				"fields":    []string{"full_name^3", "headline^2", "bio", "skills^2"},
				"fuzziness": "AUTO",
			},
		})
	}

	if len(filters.Skills) > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"terms": map[string]interface{}{
				"skills": filters.Skills,
			},
		})
	}

	if filters.Location != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"location": filters.Location,
			},
		})
	}

	if filters.ExperienceLevel != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"experience_level": filters.ExperienceLevel,
			},
		})
	}

	if filters.MinYearsOfExperience > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"years_of_experience": map[string]interface{}{
					"gte": filters.MinYearsOfExperience,
				},
			},
		})
	}

	if filters.MaxYearsOfExperience > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"years_of_experience": map[string]interface{}{
					"lte": filters.MaxYearsOfExperience,
				},
			},
		})
	}

	if filters.IsRemote != nil {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"is_remote": *filters.IsRemote,
			},
		})
	}

	if filters.IsAvailable != nil {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"is_available": *filters.IsAvailable,
			},
		})
	}

	return query
}

func BuildCompanySearchQuery(filters CompanySearchFilters) map[string]interface{} {
	from := (filters.Page - 1) * filters.Limit
	if from < 0 {
		from = 0
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	query := map[string]interface{}{
		"from": from,
		"size": filters.Limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []interface{}{},
				"filter": []interface{}{},
			},
		},
		"aggs": BuildCompanyAggregations(),
	}

	switch filters.SortBy {
	case "rating":
		query["sort"] = []interface{}{
			map[string]interface{}{"rating": "desc"},
		}
	case "total_jobs":
		query["sort"] = []interface{}{
			map[string]interface{}{"total_jobs": "desc"},
		}
	default:
		query["sort"] = []interface{}{
			map[string]interface{}{"_score": "desc"},
			map[string]interface{}{"total_jobs": "desc"},
		}
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

	if filters.Keyword != "" {
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     filters.Keyword,
				"fields":    []string{"name^3", "description^2", "industry"},
				"fuzziness": "AUTO",
			},
		})
	}

	if filters.Industry != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"industry": filters.Industry,
			},
		})
	}

	if filters.Location != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"match": map[string]interface{}{
				"location": filters.Location,
			},
		})
	}

	if filters.Size != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"size": filters.Size,
			},
		})
	}

	if filters.MinRating > 0 {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"rating": map[string]interface{}{
					"gte": filters.MinRating,
				},
			},
		})
	}

	if filters.IsVerified != nil {
		boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"is_verified": *filters.IsVerified,
			},
		})
	}

	return query
}

func BuildCandidateAggregations() map[string]interface{} {
	return map[string]interface{}{
		"by_skills": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "skills",
				"size":  30,
			},
		},
		"by_experience_level": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "experience_level",
				"size":  10,
			},
		},
		"by_location": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "location",
				"size":  20,
			},
		},
		"experience_range": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "years_of_experience",
			},
		},
	}
}

func BuildCompanyAggregations() map[string]interface{} {
	return map[string]interface{}{
		"by_industry": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "industry",
				"size":  20,
			},
		},
		"by_size": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "size",
				"size":  10,
			},
		},
		"by_location": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "location.keyword",
				"size":  20,
			},
		},
		"rating_stats": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "rating",
			},
		},
	}
}

type CandidateSearchResult struct {
	Hits []CandidateHit
	Aggs map[string]interface{}
}

type CompanySearchResult struct {
	Hits []CompanyHit
	Aggs map[string]interface{}
}

type CandidateHit struct {
	Score  float64
	Source CandidateDocument
}

type CompanyHit struct {
	Score  float64
	Source CompanyDocument
}

func (c *ESClient) SearchCandidates(ctx context.Context, filters CandidateSearchFilters) (*CandidateSearchResult, error) {
	query := BuildCandidateSearchQuery(filters)
	resp, err := c.Search(ctx, "candidates", query)
	if err != nil {
		return nil, err
	}

	result := &CandidateSearchResult{
		Hits: make([]CandidateHit, 0, len(resp.Hits.Hits)),
		Aggs: resp.Aggregations,
	}

	for _, hit := range resp.Hits.Hits {
		var doc CandidateDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}
		result.Hits = append(result.Hits, CandidateHit{
			Score:  hit.Score,
			Source: doc,
		})
	}

	return result, nil
}

func (c *ESClient) SearchCompanies(ctx context.Context, filters CompanySearchFilters) (*CompanySearchResult, error) {
	query := BuildCompanySearchQuery(filters)
	resp, err := c.Search(ctx, "companies", query)
	if err != nil {
		return nil, err
	}

	result := &CompanySearchResult{
		Hits: make([]CompanyHit, 0, len(resp.Hits.Hits)),
		Aggs: resp.Aggregations,
	}

	for _, hit := range resp.Hits.Hits {
		var doc CompanyDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}
		result.Hits = append(result.Hits, CompanyHit{
			Score:  hit.Score,
			Source: doc,
		})
	}

	return result, nil
}
