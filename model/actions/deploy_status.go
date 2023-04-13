package actions

import "strings"

type DeployStatus struct {
	val DeployStatusValue
}

type DeployStatusValue string

const (
	DeployStatusPublished DeployStatusValue = "PUBLISHED"
	DeployStatusDeployed  DeployStatusValue = "DEPLOYED"
	DeployStatusUnknown   DeployStatusValue = "UNKNOWN"
)

func NewDeployStatus(value DeployStatusValue) DeployStatus {
	return DeployStatus{val: value}
}

// IsUnknown returns false for all known variants of DeployStatus and true otherwise.
func (ds DeployStatus) IsUnknown() bool {
	switch ds.val {
	case DeployStatusPublished, DeployStatusDeployed:
		return false
	}
	return true
}

func (ds DeployStatus) Value() DeployStatusValue {
	if ds.IsUnknown() {
		return DeployStatusUnknown
	}
	return ds.val
}

func (ds DeployStatus) String() string {
	return string(ds.val)
}

func (ds DeployStatus) MarshalText() ([]byte, error) {
	return []byte(ds.val), nil
}

func (ds *DeployStatus) UnmarshalText(data []byte) error {
	switch v := strings.ToUpper(string(data)); v {
	default:
		*ds = NewDeployStatus(DeployStatusValue(v))
	case "PUBLISHED":
		*ds = NewDeployStatus(DeployStatusPublished)
	case "DEPLOYED":
		*ds = NewDeployStatus(DeployStatusDeployed)
	}
	return nil
}
