package jira

// SearchResult represents the response from Jira's /search endpoint
type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// Issue represents a single Jira issue
type Issue struct {
	ID  string `json:"id"`
	Key string `json:"key"`
	// Jira puts most of the actual data inside a "fields" object
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	// Add Status, Assignee, etc. as needed later
}
