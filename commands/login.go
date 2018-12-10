package commands

import (
	"errors"
	"fmt"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"github.com/tenderly/tenderly-cli/userError"
	"os"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tenderly/tenderly-cli/config"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "User authentication.",
	Run: func(cmd *cobra.Command, args []string) {
		rest := newRest()

		email, err := promptEmail()
		if err != nil {
			userError.LogErrorf("prompt email failed: %s", err)
			os.Exit(1)
		}

		password, err := promptPassword()
		if err != nil {
			userError.LogErrorf("prompt password failed: %s", err)
			os.Exit(1)
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
			os.Exit(1)
		}
		if tokenResponse.Error != nil {
			userError.LogErrorf("login call: %s", tokenResponse.Error)
			os.Exit(1)
		}

		config.SetGlobalConfig(config.Token, tokenResponse.Token)

		user, err := rest.User.User()
		if err != nil {
			fmt.Printf("cannot fetch user info: %s\n", err)
			os.Exit(0)
		}

		config.SetGlobalConfig("account_id", user.ID)

		err = config.WriteGlobalConfig()
		if err != nil {
			userError.LogErrorf(
				"login call: write global config: %s",
				userError.NewUserError(err, "Couldn't write global config file"),
			)
		}
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
