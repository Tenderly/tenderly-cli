package payloads

import (
	"github.com/tenderly/tenderly-cli/model"
	"regexp"
)

type GetProjectsResponse struct {
	Projects []*model.Project `json:"projects"`
	Error    *ApiError        `json:"error"`
}

var projectIDFormat = regexp.MustCompile("^[a-zA-Z0-9-_]{5,20}$")

type ProjectRequest struct {
	Name string `json:"name"`
}

func (r ProjectRequest) Valid() bool {
	return r.Name != "" && projectIDFormat.MatchString(r.Name)
}

type ProjectResponse struct {
	Project *model.Project `json:"project"`
	Error   *ApiError      `json:"error"`
}
