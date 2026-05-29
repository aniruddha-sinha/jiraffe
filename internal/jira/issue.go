package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Issue struct {
	ID  string `json:"id"`
	Key string `json:"key"`

	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	// Add Status, Assignee, etc. as needed later
}

type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

type IssueService struct {
	issueClient *Client
}

func NewIssueService(client *Client) *IssueService {
	return &IssueService{issueClient: client}
}

func (is *IssueService) List(ctx context.Context, projectKey string) ([]Issue, error) {
	fullURL, err := is.issueClient.getEndpointURL(urlTemplateSearchAPI, is.issueClient.creds.Org())
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	query.Add("jql", fmt.Sprintf(`project="%s"`, projectKey))
	query.Add("fields", "summary,key")
	parsedURL.RawQuery = query.Encode()

	request, err := is.issueClient.NewRequest(ctx, http.MethodGet, parsedURL.String())
	if err != nil {
		return nil, err
	}

	response, err := is.issueClient.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err := mapStatusToError(response.StatusCode); err != nil {
		return nil, err
	}

	var searchResult SearchResult
	if err := json.NewDecoder(response.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResult.Issues, nil
}
