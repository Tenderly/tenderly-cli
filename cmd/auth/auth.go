package auth

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
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

	viper.Set("token", token.Token)

	// TODO we cpan probably extract username from token
	user, err := rest.User.User()
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to fetch user: %s", err))
		os.Exit(0)
	}

	config.SetRC("organisation", user.Username)
	viper.Set("organisation", user.Username)
	viper.WriteConfig()
	config.WriteRC()
}

func promptEmail() (string, error) {
	validate := func(input string) error {
		re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		if !re.MatchString(input) {
			return errors.New("email not valid")
		}
		return nil
	}

	promptEmail := promptui.Prompt{
		Label:    "Enter your email",
		Validate: validate,
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
