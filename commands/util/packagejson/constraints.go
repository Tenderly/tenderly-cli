package packagejson

import "github.com/hashicorp/go-version"

type Constraints map[string]string

func (dc Constraints) findConstraints(dependencyName string) (version.Constraints, error) {
	constraintString, ok := dc[dependencyName]
	if !ok {
		return nil, nil
	}

	constraints, err := version.NewConstraint(constraintString)
	if err != nil {
		return nil, err
	}

	return constraints, nil
}

var runtimesToConstraints = map[string]*Constraints{
	"V1": {
		"axios":             "<1.0.0",
		"@tenderly/actions": "<0.1.0",
	},
	"V2": {
		"axios":             "<1.0.0",
		"@tenderly/actions": ">=0.1.0",
	},
}
