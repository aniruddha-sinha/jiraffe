package jira

import (
	"errors"
	"net/mail"
)

var ErrInvalidEmail = errors.New("invalid email address")

type profile struct {
	email string
	org   string
}

// NewProfile is the only way to construct a Profile.
// It guarantees the returned Profile is perfectly valid.
func NewProfile(email, org string) (profile, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return profile{}, ErrInvalidEmail
	}

	return profile{
		email: email,
		org:   org,
	}, nil
}

func (p profile) Email() string {
	return p.email
}

func (p profile) Org() string {
	return p.org
}
