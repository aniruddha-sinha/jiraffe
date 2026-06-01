package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"`
	Status      IssueStatus     `json:"status"`
	Priority    IssuePriority   `json:"priority"`
	Assignee    *IssueUser      `json:"assignee"`
	IssueType   IssueType       `json:"issueType"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
}

type IssueStatus struct {
	Name string `json:"name"`
}

type IssuePriority struct {
	Name string `json:"name"`
}

type IssueUser struct {
	DisplayName string `json:"displayName"`
}

type IssueType struct {
	Name string `json:"name"`
}

type adfDocument struct {
	Content []adfBlock `json:"content"`
}

type adfBlock struct {
	Content []adfInline `json:"content"`
}

type adfInline struct {
	Text string `json:"text"`
}

type SearchResult struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken"`
}

type IssueService struct {
	issueClient *Client
}

func NewIssueService(client *Client) *IssueService {
	return &IssueService{issueClient: client}
}

func (i *Issue) Summary() string {
	return i.Fields.Summary
}

func (i *Issue) Status() string {
	return i.Fields.Status.Name
}

func (i *Issue) Priority() string {
	return i.Fields.Priority.Name
}

func (i *Issue) Assignee() string {
	if i.Fields.Assignee == nil {
		return "unassigned"
	}

	return i.Fields.Assignee.DisplayName
}

func (i *Issue) Type() string {
	return i.Fields.IssueType.Name
}

func (i *Issue) Description() string {
	if len(i.Fields.Description) == 0 || string(i.Fields.Description) == "null" {
		return ""
	}

	var s string
	if err := json.Unmarshal(i.Fields.Description, &s); err == nil {
		return s
	}

	var doc adfDocument
	if err := json.Unmarshal(i.Fields.Description, &doc); err == nil {
		return extractADFText(doc)
	}

	return ""
}

func extractADFText(doc adfDocument) string {
	var parts []string
	for _, block := range doc.Content {
		for _, inline := range block.Content {
			if inline.Text != "" {
				parts = append(parts, inline.Text)
			}
		}
	}

	return strings.Join(parts, "\n")
}

func (i *Issue) Created() string {
	return i.Fields.Created
}

func (i *Issue) Updated() string {
	return i.Fields.Updated
}

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

	response, err := is.issueClient.Do(ctx, http.MethodGet, reqURL)
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
	query.Add("fields", "summary,status,priority,assignee,issueType")
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

	response, err := is.issueClient.Do(ctx, http.MethodGet, reqURL)
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
