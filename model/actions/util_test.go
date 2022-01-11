package actions_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/model/actions"
)

func MustReadTest(filename string) []byte {
	_, thisFilename, _, _ := runtime.Caller(0)
	path := strings.TrimSuffix(thisFilename, "util_test.go") + fmt.Sprintf("yaml/%s.yaml", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("read test case %s", filename)))
	}
	return content
}

func MustReadTrigger(filename string) (actions.Trigger, actions.ValidateResponse, bool) {
	testCase := MustReadTest(filename)

	var trigger actions.Trigger
	err := yaml.Unmarshal(testCase, &trigger)
	if err != nil {
		panic(errors.Wrap(err, "unmarshal trigger"))
	}

	validatorResponse := trigger.Validate("test")
	valid := len(validatorResponse.Errors) == 0

	return trigger, validatorResponse, valid
}

func MustReadTriggerAndValidate(filename string) actions.Trigger {
	trigger, response, ok := MustReadTrigger(filename)
	if !ok {
		for _, e := range response.Errors {
			fmt.Println(e)
		}
		panic("trigger validation failed")
	}
	return trigger
}

func MustReadTriggerAndFailValidate(filename string) actions.Trigger {
	trigger, _, ok := MustReadTrigger(filename)
	if ok {
		panic("trigger validation did not fail but expected")
	}
	return trigger
}
