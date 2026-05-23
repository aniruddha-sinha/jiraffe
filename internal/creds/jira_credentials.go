package creds

import (
	"fmt"

	"github.com/aniruddha-sinha/jiraffe/internal/config"
)

var (
	JiraConfigEmailKey        = "auth.jira.email"
	JiraConfigOrgKey          = "auth.jira.org"
	JiraConfigEncodedTokenKey = "auth.jira.encoded_token" // nolint:gosec // this is a config key and not an actual token
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

func (jc *JiraCreds) Store() error {
	// 1. Handle Email
	storedEmail, err := config.Cfg.Get(JiraConfigEmailKey)
	if err != nil {
		if err := config.Cfg.Upsert(JiraConfigEmailKey, jc.email); err != nil {
			return err
		}
	} else if storedEmail != jc.email {
		fmt.Println("stored email different from input email")
		if err := config.Cfg.Upsert(JiraConfigEmailKey, jc.email); err != nil {
			return err
		}
	}

	// 2. Handle Org
	storedOrg, err := config.Cfg.Get(JiraConfigOrgKey)
	if err != nil {
		if err := config.Cfg.Upsert(JiraConfigOrgKey, jc.org); err != nil {
			return err
		}
	} else if storedOrg != jc.org {
		fmt.Println("stored org is different from input org")
		if err := config.Cfg.Upsert(JiraConfigOrgKey, jc.org); err != nil {
			return err
		}
	}

	// 3. Handle Token
	storedEncodedToken, err := config.Cfg.Get(JiraConfigEncodedTokenKey)
	if err != nil {
		if err := config.Cfg.Upsert(JiraConfigEncodedTokenKey, jc.token); err != nil {
			return err
		}
	} else if storedEncodedToken != jc.token {
		fmt.Println("stored token is different from input token")
		if err := config.Cfg.Upsert(JiraConfigEncodedTokenKey, jc.token); err != nil {
			return err
		}
	}

	fmt.Println("config saved!")
	return nil
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
