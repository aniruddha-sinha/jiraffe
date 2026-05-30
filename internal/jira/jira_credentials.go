package jira

import (
	"context"
	"fmt"
)

type JiraCreds struct {
	email string
	org   string
	token string
}

func NewJiraCreds(email, org, token string) *JiraCreds {
	return &JiraCreds{
		email: email,
		org:   org,
		token: token,
	}
}

func (jc *JiraCreds) Email() string {
	return jc.email
}

func (jc *JiraCreds) Org() string {
	return jc.org
}

func (jc *JiraCreds) EncodedAPIToken() string {
	return jc.token
}

func (jc *JiraCreds) EnsureAuthentication(ctx context.Context) error {
	client := NewClient(jc)
	fullURL, err := client.BuildURL(urlTemplateValidateMyselfAPI, apiVersion)
	if err != nil {
		return err
	}

	if err := client.validateToken(ctx, fullURL); err != nil {
		return fmt.Errorf("%w:%w", ErrAPITokenValidityVerification, err)
	}
	return nil
}
