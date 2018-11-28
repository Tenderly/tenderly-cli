package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/rest"
	"github.com/tenderly/tenderly-cli/rest/call"
)

func Start(rest rest.Rest) {
	email, err := promptEmail()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	password, err := promptPassword()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	token, err := rest.Auth.Login(call.LoginRequest{
		Username: email,
		Password: password,
	})

	if err != nil {
		fmt.Println("invalid credentials")
		os.Exit(0)
	}

	config.SetGlobalConfig("token", token.Token)

	// TODO we cpan probably extract username from token
	user, err := rest.User.User()
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to fetch user: %s", err))
		os.Exit(0)
	}

	config.SetProjectConfig("organisation", user.Username)
	config.SetGlobalConfig("organisation", user.Username)
	config.WriteGlobalConfig()
	config.WriteProjectConfig()
}

func promptEmail() (string, error) {
	promptEmail := promptui.Prompt{
		Label: "Enter your email",
		Validate: func(input string) error {
			re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
			if !re.MatchString(input) {
				return errors.New("email not valid")
			}
			return nil
		},
	}

	result, err := promptEmail.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func promptPassword() (string, error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}
