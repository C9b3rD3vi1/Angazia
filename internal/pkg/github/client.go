package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
}

type GitHubUser struct {
	ID          int        `json:"id"`
	Login       string     `json:"login"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	AvatarURL   string     `json:"avatar_url"`
	HTMLURL     string     `json:"html_url"`
	Bio         string     `json:"bio"`
	Company     string     `json:"company"`
	Location    string     `json:"location"`
	Blog        string     `json:"blog"`
	CreatedAt   *time.Time `json:"created_at"`
	PublicRepos int        `json:"public_repos"`
	Followers   int        `json:"followers"`
	Following   int        `json:"following"`
	Hireable    bool       `json:"hireable"`
	Twitter     string     `json:"twitter_username"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

type GitHubRepository struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	Private     bool      `json:"private"`
	Fork        bool      `json:"fork"`
	Stargazers  int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	Watchers    int       `json:"watchers_count"`
	OpenIssues  int       `json:"open_issues_count"`
	Language    string    `json:"language"`
	Size        int       `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	PushedAt    time.Time `json:"pushed_at"`
	HasWiki     bool      `json:"has_wiki"`
	HasProjects bool      `json:"has_projects"`
	License     *struct {
		Name string `json:"name"`
	} `json:"license"`
}

type GitHubContribution struct {
	Date  time.Time
	Count int
}

func NewClient(accessToken string) *Client {
	return &Client{
		AccessToken: accessToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetAuthenticatedUser() (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	
	c.setHeaders(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (c *Client) GetPrimaryEmail() (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	
	c.setHeaders(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}
	
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	
	return "", fmt.Errorf("no email found")
}

func (c *Client) GetRepositories() ([]*GitHubRepository, error) {
	var allRepos []*GitHubRepository
	page := 1
	perPage := 100
	
	for {
		url := fmt.Sprintf("https://api.github.com/user/repos?sort=updated&per_page=%d&page=%d", perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		
		c.setHeaders(req)
		
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			return nil, err
		}
		
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
		}
		
		var repos []*GitHubRepository
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, err
		}
		
		if len(repos) == 0 {
			break
		}
		
		allRepos = append(allRepos, repos...)
		
		if len(repos) < perPage {
			break
		}
		
		page++
	}
	
	return allRepos, nil
}

type graphQLRequest struct {
	Query     string `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   graphQLData `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type graphQLData struct {
	Viewer viewerData `json:"viewer"`
}

type viewerData struct {
	ContributionsCollection contributionsCollection `json:"contributionsCollection"`
}

type contributionsCollection struct {
	ContributionCalendar contributionCalendar `json:"contributionCalendar"`
}

type contributionCalendar struct {
	Weeks []weekData `json:"weeks"`
}

type weekData struct {
	ContributionDays []contributionDay `json:"contributionDays"`
}

type contributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}

func (c *Client) GetContributions(days int) ([]*GitHubContribution, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -days)
	to := now

	query := `query($from: DateTime!, $to: DateTime!) {
	  viewer {
	    contributionsCollection(from: $from, to: $to) {
	      contributionCalendar {
	        weeks {
	          contributionDays {
	            date
	            contributionCount
	          }
	        }
	      }
	    }
	  }
	}`

	reqBody := graphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"from": from.Format(time.RFC3339),
			"to":   to.Format(time.RFC3339),
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub GraphQL API error: %d - %s", resp.StatusCode, string(respBytes))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GitHub GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	calendar := gqlResp.Data.Viewer.ContributionsCollection.ContributionCalendar
	contributions := make([]*GitHubContribution, 0)

	for _, week := range calendar.Weeks {
		for _, day := range week.ContributionDays {
			date, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				continue
			}
			contributions = append(contributions, &GitHubContribution{
				Date:  date,
				Count: day.ContributionCount,
			})
		}
	}

	return contributions, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Kenyan-Dev-Marketplace/1.0")
}