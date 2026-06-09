package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var urlTemplateGetSprintDetails = "/rest/agile/1.0/board/%d/sprint"

type JiraUser struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func NewJiraUser(acctID, displayName string) *JiraUser {
	return &JiraUser{
		AccountID:   acctID,
		DisplayName: displayName,
	}
}

type BoardResponse struct {
	Values []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"values"`
}

// SprintResponse matches the Jira Agile Sprint list schema
type SprintResponse struct {
	Values []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
	} `json:"values"`
}

func (c *Client) ResolveEmailToAtlassianUserID(ctx context.Context, email string) (string, error) {
	reqURL, err := c.buildURLForQueryParams(urlTemplateSearchAtlassianUserByEmail, apiVersion)
	if err != nil {
		return "", err
	}

	query := reqURL.Query()
	query.Add("query", email)
	reqURL.RawQuery = query.Encode()
	response, err := c.Do(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", err
	}

	var jiraUser []JiraUser
	if err := json.NewDecoder(response.Body).Decode(&jiraUser); err != nil {
		return "", fmt.Errorf("failed to decode response : %w", err)
	}

	return jiraUser[0].AccountID, nil
}

func (c *Client) ResolveSprintNameToID(ctx context.Context, projectKey, sprintName string) (int, error) {
	boardURL, err := c.buildURLForQueryParams("/rest/agile/1.0/board")
	if err != nil {
		return 0, err
	}
	boardQuery := boardURL.Query()
	boardQuery.Add("projectKeyOrId", projectKey)
	boardURL.RawQuery = boardQuery.Encode()

	boardResp, err := c.Do(ctx, http.MethodGet, boardURL.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch agile boards: %w", err)
	}
	defer func() { _ = boardResp.Body.Close() }()

	var boards BoardResponse
	if err := json.NewDecoder(boardResp.Body).Decode(&boards); err != nil {
		return 0, err
	}
	if len(boards.Values) == 0 {
		return 0, fmt.Errorf("no agile boards found for project %s", projectKey)
	}

	boardID := boards.Values[0].ID

	sprintURL, err := c.buildURLForQueryParams(urlTemplateGetSprintDetails, boardID)
	if err != nil {
		return 0, err
	}

	sprintResp, err := c.Do(ctx, http.MethodGet, sprintURL.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch sprints for board %w", err)
	}
	defer func() { _ = sprintResp.Body.Close() }()

	var sprints SprintResponse
	if err := json.NewDecoder(sprintResp.Body).Decode(&sprints); err != nil {
		return 0, err
	}

	for _, s := range sprints.Values {
		if strings.EqualFold(s.Name, sprintName) {
			return s.ID, nil
		}
	}

	return 0, fmt.Errorf("could not find an active or planned sprint named '%s' in project %s", sprintName, projectKey)
}
