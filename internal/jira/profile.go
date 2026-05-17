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

func NewProfile(email, org string) (*profile, error) {
	if err := validate(email); err != nil {
		return &profile{}, err
	}
	return &profile{email: email, org: org}, nil
}

func validate(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	return nil
}
