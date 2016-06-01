package commands

import (
	"errors"
	"fmt"

	"github.com/concourse/atc"
	"github.com/concourse/fly/commands/internal/displayhelpers"
	"github.com/concourse/fly/commands/internal/flaghelpers"
	"github.com/concourse/fly/rc"
	"github.com/vito/go-interact/interact"
)

type SetTeamCommand struct {
	TeamName string `short:"n" long:"team-name" required:"true"        description:"The team to create or modify"`

	BasicAuth struct {
		Username string `long:"username" description:"Username to use for basic auth."`
		Password string `long:"password" description:"Password to use for basic auth."`
	} `group:"Basic Authentication" namespace:"basic-auth"`

	GitHubAuth struct {
		ClientID      string                       `long:"client-id"     description:"Application client ID for enabling GitHub OAuth."`
		ClientSecret  string                       `long:"client-secret" description:"Application client secret for enabling GitHub OAuth."`
		Organizations []string                     `long:"organization"  description:"GitHub organization whose members will have access." value-name:"ORG"`
		Teams         []flaghelpers.GitHubTeamFlag `long:"team"          description:"GitHub team whose members will have access." value-name:"ORG/TEAM"`
		Users         []string                     `long:"user"          description:"GitHub user to permit access." value-name:"LOGIN"`
	} `group:"GitHub Authentication" namespace:"github-auth"`

	CFAuth CFAuth `group:"CF Authentication" namespace:"cf-auth"`
}

type CFAuth struct {
	ClientID     string   `long:"client-id"     description:"Application client ID for enabling UAA OAuth."`
	ClientSecret string   `long:"client-secret" description:"Application client secret for enabling UAA OAuth."`
	Spaces       []string `long:"space"         description:"Space GUID for a CF space whose developers will have access."`
	AuthURL      string   `long:"auth-url"      description:"UAA AuthURL endpoint."`
	TokenURL     string   `long:"token-url"     description:"UAA TokenURL endpoint."`
	APIURL       string   `long:"api-url"       description:"CF API endpoint."`
}

func (auth *CFAuth) IsConfigured() bool {
	return auth.ClientID != "" ||
		auth.ClientSecret != "" ||
		len(auth.Spaces) > 0 ||
		auth.AuthURL != "" ||
		auth.TokenURL != "" ||
		auth.APIURL != ""
}

func (auth *CFAuth) Validate() error {
	if auth.ClientID == "" || auth.ClientSecret == "" {
		return errors.New("Both client-id and client-secret are required for cf-auth.")
	}
	if len(auth.Spaces) == 0 {
		return errors.New("space is required for cf-auth.")
	}
	if auth.AuthURL == "" || auth.TokenURL == "" || auth.APIURL == "" {
		return errors.New("auth-url, token-url and api-url are required for cf-auth.")
	}
	return nil
}

func (command *SetTeamCommand) Execute([]string) error {
	target, err := rc.LoadTarget(Fly.Target)
	if err != nil {
		return err
	}

	err = target.Validate()
	if err != nil {
		return err
	}

	hasBasicAuth, hasGitHubAuth, err := command.ValidateFlags()
	if err != nil {
		return err
	}

	fmt.Println("Team Name:", command.TeamName)
	fmt.Println("Basic Auth:", authMethodStatusDescription(hasBasicAuth))
	fmt.Println("GitHub Auth:", authMethodStatusDescription(hasGitHubAuth))

	confirm := false
	err = interact.NewInteraction("apply configuration?").Resolve(&confirm)
	if err != nil {
		return err
	}

	if !confirm {
		displayhelpers.Failf("bailing out")
	}

	team := atc.Team{}

	if hasBasicAuth {
		team.BasicAuth = &atc.BasicAuth{
			BasicAuthUsername: command.BasicAuth.Username,
			BasicAuthPassword: command.BasicAuth.Password,
		}
	}

	if hasGitHubAuth {
		team.GitHubAuth = &atc.GitHubAuth{
			ClientID:      command.GitHubAuth.ClientID,
			ClientSecret:  command.GitHubAuth.ClientSecret,
			Organizations: command.GitHubAuth.Organizations,
			Users:         command.GitHubAuth.Users,
		}

		for _, ghTeam := range command.GitHubAuth.Teams {
			team.GitHubAuth.Teams = append(team.GitHubAuth.Teams, atc.GitHubTeam{
				OrganizationName: ghTeam.OrganizationName,
				TeamName:         ghTeam.TeamName,
			})
		}
	}

	_, _, _, err = target.Client().Team(command.TeamName).CreateOrUpdate(team)
	if err != nil {
		return err
	}

	fmt.Println("team created")
	return nil
}

func (command *SetTeamCommand) ValidateFlags() (bool, bool, error) {
	hasBasicAuth := command.BasicAuth.Username != "" || command.BasicAuth.Password != ""
	if hasBasicAuth && (command.BasicAuth.Username == "" || command.BasicAuth.Password == "") {
		return false, false, errors.New("Both username and password are required for basic auth.")
	}
	hasGitHubAuth := command.GitHubAuth.ClientID != "" || command.GitHubAuth.ClientSecret != "" ||
		len(command.GitHubAuth.Organizations) > 0 || len(command.GitHubAuth.Teams) > 0 || len(command.GitHubAuth.Users) > 0
	if hasGitHubAuth {
		if command.GitHubAuth.ClientID == "" || command.GitHubAuth.ClientSecret == "" {
			return false, false, errors.New("Both client-id and client-secret are required for github-auth.")
		}
		if len(command.GitHubAuth.Organizations) == 0 &&
			len(command.GitHubAuth.Teams) == 0 &&
			len(command.GitHubAuth.Users) == 0 {
			return false, false, errors.New("At least one of the following is required for github-auth: organizations, teams, users")
		}
	}

	if command.CFAuth.IsConfigured() {
		err := command.CFAuth.Validate()
		if err != nil {
			return false, false, err
		}
	}

	return hasBasicAuth, hasGitHubAuth, nil
}

func authMethodStatusDescription(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
