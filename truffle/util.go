package truffle

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	lettersLen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

func extractConfigWithDivider(config, divider string) (string, error) {
	reg := regexp.MustCompile(fmt.Sprintf("%s(?P<Config>.*)%s", divider, divider))
	matches := reg.FindStringSubmatch(config)

	if len(matches) < 2 {
		return "", errors.New("couldn't extract config with divider")
	}

	return matches[1], nil
}
