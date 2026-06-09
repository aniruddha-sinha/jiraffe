package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Field struct {
	ID   string
	Name string
}

func NewField(id, name string) *Field {
	return &Field{
		ID:   id,
		Name: name,
	}
}

type FieldService struct {
	fieldClient *Client
}

func NewFieldService(client *Client) *FieldService {
	return &FieldService{fieldClient: client}
}

func (fs *FieldService) GetAll(ctx context.Context) ([]Field, error) {
	reqURL, err := fs.fieldClient.buildRawURL(urlTemplateGetFields, apiVersion)
	if err != nil {
		return nil, err
	}

	response, err := fs.fieldClient.Do(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var fields []Field
	if err := json.NewDecoder(response.Body).Decode(&fields); err != nil {
		return nil, fmt.Errorf("failed to decode response : %w", err)
	}

	return fields, nil
}
