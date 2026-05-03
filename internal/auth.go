package internal

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
func (p UserProfile) HandleAuthentication() error {
	return nil
}
