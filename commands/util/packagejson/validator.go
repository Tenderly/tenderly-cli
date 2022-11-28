package packagejson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/tenderly/tenderly-cli/userError"
)

type Validator struct {
	constraints *Constraints
}

func NewValidator(runtimeName string) *Validator {
	constraints := runtimesToConstraints[strings.ToUpper(runtimeName)]
	if constraints == nil {
		return &Validator{
			constraints: &Constraints{},
		}
	}

	return &Validator{
		constraints: constraints,
	}
}

type ValidationError struct {
	Name                 string
	PackageJsonVersion   string
	Constraint           string
	VersionToBeInstalled string
}

type ValidationResult struct {
	Success bool
	Errors  []*ValidationError
}

func (dv *Validator) Validate(dependencies map[string]string) (*ValidationResult, error) {
	var validationErrors []*ValidationError

	for packageName, packageVersion := range dependencies {
		constraint, err := dv.constraints.findConstraints(packageName)
		if err != nil {
			return nil, err
		}

		if constraint == nil {
			continue
		}

		versionToBeInstalled, err := FindPackageVersion(packageName, packageVersion)
		if err != nil {
			return nil, err
		}

		parsedVersion, err := version.NewVersion(versionToBeInstalled)
		if err != nil {
			return nil, errors.New("error parsing version")
		}

		if !constraint.Check(parsedVersion) {
			validationErrors = append(validationErrors, &ValidationError{
				Name:                 packageName,
				PackageJsonVersion:   packageVersion,
				Constraint:           constraint.String(),
				VersionToBeInstalled: versionToBeInstalled,
			})
		}
	}

	return &ValidationResult{
		Success: len(validationErrors) == 0,
		Errors:  validationErrors,
	}, nil
}

func FindPackageVersion(name string, version string) (string, error) {
	lookup := name + "@" + version

	cmd := exec.Command("npm", "view", lookup, "version", "--json")
	var out bytes.Buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = &out

	err := cmd.Start()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to run: npm view %s version --json", lookup))
	}

	err = cmd.Wait()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to run: npm view %s version --json", lookup))
	}

	ver, err := unmarshalVersion(out)
	if err != nil {
		return "", err
	}

	return ver, nil
}

func unmarshalVersion(message bytes.Buffer) (string, error) {
	var rawMessage json.RawMessage

	err := json.Unmarshal(message.Bytes(), &rawMessage)
	if err != nil {
		return "", userError.NewUserError(err, "error unmarshalling response from npm")
	}

	switch rawMessage[0] {
	case '"':
		var v string
		err = json.Unmarshal(rawMessage, &v)
		if err != nil {
			return "", userError.NewUserError(err, "error unmarshalling response from npm")
		}

		return v, nil
	case '[':
		var versions []string
		message = *bytes.NewBufferString("aha")
		err = json.Unmarshal(rawMessage, &versions)
		if err != nil {
			return "", userError.NewUserError(err, "error unmarshalling response from npm")
		}

		highestVersion := versions[len(versions)-1]
		return highestVersion, nil
	default:
		err := errors.New(fmt.Sprintf("Unexpected response from npm: %s", rawMessage))
		return "", userError.NewUserError(err, "error unmarshalling response from npm")
	}
}
