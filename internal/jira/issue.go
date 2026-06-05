package jira

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IssueService struct {
	issueClient *Client
}

func NewIssueService(client *Client) *IssueService {
	return &IssueService{issueClient: client}
}

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
	Version int        `json:"version"`
	Type    string     `json:"type"`
	Content []adfBlock `json:"content"`
}

type adfBlock struct {
	Type    string      `json:"type"`
	Content []adfInline `json:"content"`
}

type adfInline struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SearchResult struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken"`
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

func (i *Issue) Description() (string, error) {
	rawDescripton := i.Fields.Description
	if len(rawDescripton) == 0 || string(rawDescripton) == "null" {
		return "", nil
	}

	var adfDoc adfDocument
	if err := json.Unmarshal(rawDescripton, &adfDoc); err != nil {
		return "", fmt.Errorf("failed to decode json string into ADF document: %w", err)
	}

	return extractADFText(adfDoc), nil
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

func (i *Issue) Json() (string, error) {
	jsonOut, err := json.MarshalIndent(i, "", "	")
	if err != nil {
		return "", err
	}

	return string(jsonOut), nil
}

func (i *Issue) String() (string, error) {
	desc, err := i.Description()
	if err != nil {
		return "", err
	}
	prettyPrintIssue := fmt.Sprintf(
		`%s
				Summary:	%s
				Type:		%s
				Status:		%s
				Priority:	%s
				Assignee:	%s
				Created:	%s
				Updated:	%s
				Description:	%s
				`,
		i.Key, i.Summary(), i.Type(), i.Status(), i.Priority(), i.Assignee(), i.Created(), i.Updated(), desc)

	return prettyPrintIssue, nil
}
