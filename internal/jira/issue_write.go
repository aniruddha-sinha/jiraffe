package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateIssueRequest struct {
	Fields CreateIssueFields `json:"fields"`
}

func NewCreateIssueRequest(createIssueFields *CreateIssueFields) *CreateIssueRequest {
	return &CreateIssueRequest{Fields: *createIssueFields}
}

type CreateIssueFields struct {
	Project     ProjectRef   `json:"project"`
	Summary     string       `json:"summary"`
	Description *adfDocument `json:"description,omitempty"`
	IssueType   IssueTypeRef `json:"issuetype"`
	Labels      []string     `json:"labels,omitempty"`
	Assignee    *UserRef     `json:"assignee,omitempty"`
	Reporter    *UserRef     `json:"reporter,omitempty"`
	Parent      *ParentRef   `json:"parent,omitempty"`
	/**
	* a note about custom fields
	* customField has basically two values, in this case team and sprint
	* the sprint and team is reference by the json field `json:"customfield_10020,omitempty"`
	* where in customfield_10020, 10020 -> sprint id
	**/
	CustomFields map[string]any `json:"-"`
}

func NewCreateIssueFields(project ProjectRef, summary string, desc *adfDocument, issueType IssueTypeRef, labels []string, customFields map[string]any) *CreateIssueFields {
	return &CreateIssueFields{
		Project:      project,
		Summary:      summary,
		Description:  desc,
		IssueType:    issueType,
		Labels:       labels,
		CustomFields: customFields,
	}
}

type ProjectRef struct {
	Key string `json:"key"`
}

func NewProjectRef(key string) *ProjectRef {
	return &ProjectRef{
		Key: key,
	}
}

type IssueTypeRef struct {
	Name string `json:"name"`
}

func NewIssueTypeRef(name string) *IssueTypeRef {
	return &IssueTypeRef{
		Name: name,
	}
}

type UserRef struct {
	ID string `json:"id"`
}

func NewUserRef(userId string) *UserRef {
	return &UserRef{
		ID: userId,
	}
}

type ParentRef struct {
	Key string `json:"Key"`
}

func NewParentRef(parentKey string) *ParentRef {
	return &ParentRef{
		Key: parentKey,
	}
}

type IssueCreateResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

func (icr *IssueCreateResponse) PrintJSON() (string, error) {
	jsonFormatted, err := json.MarshalIndent(icr, "", " ")
	if err != nil {
		return "", err
	}

	return string(jsonFormatted), nil
}

func BuildADFDescription(text string) *adfDocument {
	if text == "" {
		return nil
	}

	return &adfDocument{
		Version: 1,
		Type:    "doc",
		Content: []adfBlock{
			{
				Type: "paragraph",
				Content: []adfInline{
					{
						Type: "text",
						Text: text,
					},
				},
			},
		},
	}
}

func (is *IssueService) Create(ctx context.Context, payload CreateIssueRequest) (*IssueCreateResponse, error) {
	reqURL, err := is.issueClient.buildRawURL(urlTemplateIssueCreate, apiVersion)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := is.issueClient.Do(ctx, http.MethodPost, reqURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	defer func() { _ = response.Body.Close() }()

	var issueCreateResponse IssueCreateResponse
	if err := json.NewDecoder(response.Body).Decode(&issueCreateResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response : %w", err)
	}

	return &issueCreateResponse, nil
}

// custom json marshaller for incorporating customFields
func (c CreateIssueFields) MarshalJSON() ([]byte, error) {
	type Alias CreateIssueFields
	baseStruct := (Alias)(c)

	baseBytes, err := json.Marshal(baseStruct)
	if err != nil {
		return nil, err
	}

	if len(c.CustomFields) == 0 {
		return baseBytes, nil
	}

	var payloadMap map[string]any
	if err := json.Unmarshal(baseBytes, &payloadMap); err != nil {
		return nil, err
	}

	for key, value := range c.CustomFields {
		payloadMap[key] = value
	}

	return json.Marshal(payloadMap)
}
