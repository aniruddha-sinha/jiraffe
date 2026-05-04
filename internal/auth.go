package internal

import (
	"encoding/base64"
	"fmt"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/term"
)

/**
** flow
** //? Accept User Profile <email, org>
** //? search for config file in ~/.config/jiraffe/config.yaml
** //? If config.yaml found
**		//? Validate email/token and move on
** //? else
** //! Prompt the user to head to the suggested URL and generate an API token
** //! once generated; copy to clipboard and paste on the terminal
** //? config gets saved
**/

func getLocalToken(p UserProfile) (bool, error) {
	if err := loadProfileConfig(); err != nil {
		return false, err
	}

	isTokenValid, err := IsTokenValid(p, viper.GetString("auth.encoded_token"))
	if err != nil {
		return false, err
	}
	return isTokenValid, nil
}

func obtainEncodedTokenFromUser(p UserProfile) (string, error) {
	fmt.Printf("\nStarting authentication for %s at %s.atlassian.net\n", p.Email, p.Org)
	fmt.Println("Click on this link to generate API token: https://id.atlassian.com/manage-profile/security/api-tokens")

	fmt.Print("Enter API token: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("\nfailed to read token: %w", err)
	}
	token := strings.TrimSpace(string(bytePassword))
	fmt.Println("\n\nVerifying credentials...")

	// 3. Encode and verify
	authStr := fmt.Sprintf("%s:%s", p.Email, token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
	return encodedAuth, nil
}

func getWebToken(p UserProfile) (bool, error) {
	encodedToken, err := obtainEncodedTokenFromUser(p)
	if err != nil {
		return false, err
	}

	isValid, err := IsTokenValid(p, encodedToken)
	if err != nil {
		return false, fmt.Errorf("verification failed: %w", err)
	}

	if isValid {
		filePath, err := Save(p, encodedToken)
		if err != nil {
			return false, err
		} else {
			fmt.Printf("Save success : %s\n", filePath)
		}

		return true, nil
	}

	return false, nil
}

func getClient(p UserProfile) (bool, error) {
	/**
	** if loadProfileConfig does not fail then validity of token is checked and
	** auth gets successful
	** if it does then the user will be asked to obtain token from web
	**/
	err := loadProfileConfig()
	if err == nil {
		return getLocalToken(p)
	} else {
		return getWebToken(p)
	}
}

func (p UserProfile) HandleAuthentication() (bool, error) {
	isAuthenticated, err := getClient(p)
	if err != nil {
		return false, err
	}
	return isAuthenticated, nil
}
