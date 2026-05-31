package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Project struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ProjectTypeKey string `json:"projectTypeKey"`
	Style          string `json:"style"`
}

type ProjectService struct {
	projectClient *Client
}

func NewProjectService(client *Client) *ProjectService {
	return &ProjectService{
		projectClient: client,
	}
}

func (p *Project) ProjectID() string {
	return p.ID
}

func (p *Project) ProjectKey() string {
	return p.Key
}

func (p *Project) ProjectName() string {
	return p.Name
}

func (p *Project) JiraProjectTypeKey() string {
	return p.ProjectTypeKey
}

func (p *Project) ProjectStyle() string {
	return p.Style
}

func (ps *ProjectService) List(ctx context.Context) ([]Project, error) {
	fullURL, err := ps.projectClient.buildRawURL(urlTemplateListProjects, apiVersion)
	if err != nil {
		return nil, err
	}

	request, err := ps.projectClient.NewRequest(ctx, http.MethodGet, fullURL)
	if err != nil {
		return nil, err
	}

	response, err := ps.projectClient.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err := mapStatusToError(response.StatusCode); err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.NewDecoder(response.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to get projects list : %w", err)
	}

	return projects, nil
}

func (ps *ProjectService) Get(ctx context.Context, projectKey string) (*Project, error) {
	fullURL, err := ps.projectClient.buildRawURL(urlTemplateProjectSearch, apiVersion, projectKey)
	if err != nil {
		return nil, err
	}

	request, err := ps.projectClient.NewRequest(ctx, http.MethodGet, fullURL)
	if err != nil {
		return &Project{}, err
	}

	response, err := ps.projectClient.httpClient.Do(request)
	if err != nil {
		return &Project{}, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err := mapStatusToError(response.StatusCode); err != nil {
		return &Project{}, err
	}

	var project Project
	if err := json.NewDecoder(response.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to get projects list : %w", err)
	}

	return &project, nil
}
