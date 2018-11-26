package whoami

import (
	"fmt"
	"os"

	"github.com/tenderly/tenderly-cli/rest"
)

func Start(rest rest.Rest) {
	user, err := rest.User.User()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Println(fmt.Sprintf(user.Email))
}
