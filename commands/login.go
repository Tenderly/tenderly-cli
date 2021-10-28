package commands

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
)

const (
	numberOfTries = 3
)

var providedEmail string
var providedPassword string
var providedAccessKey string
var providedAuthenticationMethod string
var forceLogin bool

func init() {
	loginCmd.PersistentFlags().StringVar(&providedEmail, "email", "", "The email used for logging in.")
	loginCmd.PersistentFlags().StringVar(&providedPassword, "password", "", "The password used for logging in.")
	loginCmd.PersistentFlags().StringVar(&providedAccessKey, "access-key", "", "The access key generated in your Tenderly dashboard.")
	loginCmd.PersistentFlags().StringVar(&providedAuthenticationMethod, "authentication-method", "", "Pick the authentication method. Possible values are email or access-key.")
	loginCmd.PersistentFlags().BoolVar(&forceLogin, "force", false, "Don't check if you are already logged in.")
	RootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "User authentication",
	Run: func(cmd *cobra.Command, args []string) {
		if config.IsLoggedIn() && !forceLogin {
			alreadyLoggedIn := config.GetString(config.Username)
			if len(alreadyLoggedIn) == 0 {
				alreadyLoggedIn = config.GetString(config.Email)
			}
			if len(alreadyLoggedIn) == 0 {
				alreadyLoggedIn = config.GetString(config.OrganizationName)
			}
			if len(alreadyLoggedIn) == 0 {
				alreadyLoggedIn = config.GetString(config.AccountID)
			}

			logrus.Info(Colorizer.Sprintf("It seems that you are already logged in with the account %s. "+
				"If this is not you or you want to login with a different account rerun this command with the %s flag.",
				Colorizer.Bold(Colorizer.Green(alreadyLoggedIn)),
				Colorizer.Bold(Colorizer.Green("--force")),
			))
			os.Exit(0)
		}

		if len(providedAuthenticationMethod) == 0 {
			promptAuthenticationMethod()
		}

		rest := NewRest()
		var key string
		var keyId string

		if providedAuthenticationMethod == "email" {
			key, keyId = emailLogin(rest)
		} else if providedAuthenticationMethod == "access-key" {
			key = accessKeyLogin()
		} else {
			userError.LogErrorf("unsupported authentication method: %s", userError.NewUserError(
				fmt.Errorf("non-supported authentication method: %s", providedAuthenticationMethod),
				Colorizer.Sprintf(
					"The %s can either be %s or %s",
					Colorizer.Bold(Colorizer.Green("--authentication-method")),
					Colorizer.Bold(Colorizer.Green("email")),
					Colorizer.Bold(Colorizer.Green("access-key")),
				),
			))
			os.Exit(1)
		}

		if key == "" {
			os.Exit(1)
		}

		config.SetGlobalConfig(config.AccessKey, key)
		config.SetGlobalConfig(config.AccessKeyId, keyId)

		principal, err := rest.User.Principal()
		if err != nil {
			if providedAuthenticationMethod == "access-key" {
				userError.LogErrorf("cannot fetch user info: %s", userError.NewUserError(
					err,
					fmt.Sprintf("%s", Colorizer.Red("Couldn't fetch user information. This can happen if your access key is not valid. Please try again.")),
				))
				os.Exit(1)
			}

			userError.LogErrorf("cannot fetch user info: %s", userError.NewUserError(
				err,
				"Couldn't fetch user information. Please try again. This can happen if you are running an older version of the Tenderly CLI.",
			))

			CheckVersion(true, true)

			os.Exit(1)
		}

		config.SetGlobalConfig(config.AccountID, principal.ID)

		if principal.Type == model.UserPrincipalType {
			config.SetGlobalConfig(config.Email, principal.User.Email)
		}

		if principal.Type == model.OrganizationPrincipalType {
			config.SetGlobalConfig(config.OrganizationName, principal.Organization.Name)
		}

		config.SetGlobalConfig(config.Username, principal.Username)

		WriteGlobalConfig()

		DetectedProjectMessage(
			true,
			"initialize",
			"cd %s; tenderly init",
		)
	},
}

func promptAuthenticationMethod() {
	promptLoginWith := promptui.Select{
		Label: "Select authentication method",
		Items: []string{
			"Email",
			Colorizer.Sprintf(
				"Access key can be generated at %s",
				Colorizer.Bold(Colorizer.Green("https://dashboard.tenderly.co/account/authorization")),
			),
		},
	}

	index, _, err := promptLoginWith.Run()
	if err != nil {
		userError.LogErrorf("prompt authentication method failed: %s", err)
		os.Exit(1)
	}

	providedAuthenticationMethod = "email"

	if index == 1 {
		providedAuthenticationMethod = "access-key"
	}
}

func emailLogin(rest *rest.Rest) (string, string) {
	var token string
	var tokenId string

	for i := 0; i < numberOfTries; i++ {
		var email string
		var password string
		var err error

		if providedEmail == "" {
			email, err = promptEmail()
			if err != nil {
				userError.LogErrorf("prompt email failed: %s", err)
				os.Exit(1)
			}
		} else {
			email = providedEmail
		}

		if providedPassword == "" {
			password, err = promptPassword()
			if err != nil {
				userError.LogErrorf("prompt password failed: %s", err)
				os.Exit(1)
			}
		} else {
			password = providedPassword
		}

		tokenResponse, err := rest.Auth.Login(payloads.LoginRequest{
			Email:    email,
			Password: password,
		})

		if err != nil {
			userError.LogErrorf("login call: %s", userError.NewUserError(
				err,
				"Couldn't make the login request. Please try again.",
			))
			continue
		}
		if tokenResponse.Error != nil {
			userError.LogErrorf("login call: %s", tokenResponse.Error)
			if providedEmail != "" && providedPassword != "" {
				break
			}
			continue
		}

		token = tokenResponse.Token
		tokenId = tokenResponse.ID

		break
	}

	return token, tokenId
}

func accessKeyLogin() string {
	if len(providedAccessKey) != 0 {
		return providedAccessKey
	}

	result, err := promptAccessKey()
	if err != nil {
		return ""
	}

	return result
}

func promptEmail() (string, error) {
	promptEmail := promptui.Prompt{
		Label: "Enter your email",
		Validate: func(input string) error {
			re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
			if !re.MatchString(input) {
				return errors.New("Please enter a valid e-mail address")
			}
			return nil
		},
	}

	result, err := promptEmail.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func promptPassword() (string, error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Please enter your password")
			}
			return nil
		},
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func promptAccessKey() (string, error) {
	prompt := promptui.Prompt{
		Label: "Access key",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("Please enter your access key")
			}
			return nil
		},
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}
