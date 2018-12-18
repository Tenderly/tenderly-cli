package commands

import (
	"errors"
	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
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

var providedUsername string
var providedPassword string
var forceLogin bool

func init() {
	loginCmd.PersistentFlags().StringVar(&providedUsername, "username", "", "The username used for logging in.")
	loginCmd.PersistentFlags().StringVar(&providedPassword, "password", "", "The password used for logging in.")
	loginCmd.PersistentFlags().BoolVar(&forceLogin, "force", false, "Don't check if you are already logged in.")
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "User authentication.",
	Run: func(cmd *cobra.Command, args []string) {
		if config.IsLoggedIn() && !forceLogin {
			logrus.Info(aurora.Sprintf("It seems that you are already logged in with the account %s. "+
				"If this is not you or you want to login with a different account rerun this command with the %s flag.",
				aurora.Bold(aurora.Green(config.GetString(config.AccountID))),
				aurora.Bold(aurora.Green("--force")),
			))
			os.Exit(0)
		}

		rest := newRest()
		var token string

		for i := 0; i < numberOfTries; i++ {
			var email string
			var password string
			var err error

			if providedUsername == "" {
				email, err = promptEmail()
				if err != nil {
					userError.LogErrorf("prompt email failed: %s", err)
					os.Exit(1)
				}
			} else {
				email = providedUsername
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
				Username: email,
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
				if providedUsername != "" && providedPassword != "" {
					break
				}
				continue
			}

			token = tokenResponse.Token
			break
		}

		if token == "" {
			os.Exit(1)
		}

		config.SetGlobalConfig(config.Token, token)

		user, err := rest.User.User()
		if err != nil {
			userError.LogErrorf("cannot fetch user info: %s", userError.NewUserError(
				err,
				"Couldn't fetch user information. Please try again.",
			))
			os.Exit(1)
		}

		config.SetGlobalConfig("account_id", user.ID)

		WriteGlobalConfig()

		DetectedProjectMessage(
			true,
			"initialize",
			"cd %s; tenderly init",
		)
	},
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
