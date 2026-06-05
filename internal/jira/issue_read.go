package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (is *IssueService) List(ctx context.Context, projectKey string, maxPages int) ([]Issue, error) {
	var allIssues []Issue
	nextPageToken := ""

	for page := 1; page <= maxPages; page++ {
		result, err := is.fetchPage(ctx, projectKey, nextPageToken)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, result.Issues...)
		if result.NextPageToken == "" {
			break
		}

		nextPageToken = result.NextPageToken
	}

	return allIssues, nil
}

func (is *IssueService) fetchPage(ctx context.Context, projectKey, pageToken string) (*SearchResult, error) {
	reqURL, err := is.buildSearchURL(projectKey, pageToken)
	if err != nil {
		return nil, err
	}

	response, err := is.issueClient.Do(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = response.Body.Close() }()
	var searchResult SearchResult

	if err := json.NewDecoder(response.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response : %w", err)
	}

	return &searchResult, nil
}

func (is *IssueService) buildSearchURL(projectKey, pageToken string) (string, error) {
	fullURL, err := is.issueClient.buildURLForQueryParams(urlTemplateIssueSearchAPIJQL, apiVersion)
	if err != nil {
		return "", err
	}

	query := fullURL.Query()
	query.Add("jql", fmt.Sprintf(`project=%s`, projectKey))
	query.Add("fields", "summary,status,priority,assignee,issueType,description,created,updated")
	query.Add("maxResults", "50") // need to reconsider; more results -> heavier slice
	if pageToken != "" {
		query.Add("nextPageToken", pageToken)
	}

	fullURL.RawQuery = query.Encode()

	return fullURL.String(), nil
}

func (is *IssueService) Get(ctx context.Context, issueKey string) (*Issue, error) {
	reqURL, err := is.issueClient.buildRawURL(urlTemplateIssueGet, apiVersion, issueKey)
	if err != nil {
		return nil, err
	}

	response, err := is.issueClient.Do(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = response.Body.Close() }()

	var issue Issue
	if err := json.NewDecoder(response.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response : %w", err)
	}

	return &issue, nil
}
