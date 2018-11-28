package call

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
)

var projectIDFormat = regexp.MustCompile("^[a-zA-Z0-9-_]{5,20}$")

type ProjectRequest struct {
	Name string `json:"name"`
}

type ProjectCalls struct {
}

func NewProjectCalls() *ProjectCalls {
	return &ProjectCalls{}
}

func (r ProjectRequest) Valid() bool {
	return r.Name != "" && projectIDFormat.MatchString(r.Name)
}

func (rest *ProjectCalls) CreateProject(request ProjectRequest) (*model.Project, error) {
	projectJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var project model.Project

	response := client.Request(
		"POST",
		"api/v1/account/"+config.GetString("organisation")+"/project",
		config.GetString("token"),
		bytes.NewBuffer(projectJson))
	err = json.NewDecoder(response).Decode(&project)
	return &project, err
}

func (rest *ProjectCalls) GetProject(organisationId, id string) (*model.Project, error) {
	var project *model.Project
	response := client.Request(
		"GET",
		"api/v1/account/"+organisationId+"/project/"+id,
		"",
		nil)

	err := json.NewDecoder(response).Decode(project)
	return project, err
}

func (rest *ProjectCalls) GetProjects(organisationId string) ([]*model.Project, error) {
	var projects []*model.Project
	response := client.Request(
		"GET",
		"api/v1/account/"+organisationId+"/projects",
		config.GetString("token"),
		nil)

	err := json.NewDecoder(response).Decode(&projects)
	return projects, err
}
