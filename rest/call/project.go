package call

import (
	"encoding/json"
	"io/ioutil"
	"regexp"

	"github.com/sirupsen/logrus"
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
		"api/v1/account/"+config.GetString(config.AccountID)+"/project",
		projectJson,
	)
	err = json.NewDecoder(response).Decode(&project)
	return &project, err
}

func (rest *ProjectCalls) GetProject(accountId, id string) (*model.Project, error) {
	var project *model.Project
	response := client.Request(
		"GET",
		"api/v1/account/"+accountId+"/project/"+id,
		nil,
	)

	err := json.NewDecoder(response).Decode(project)
	return project, err
}

func (rest *ProjectCalls) GetProjects(accountId string) ([]*model.Project, error) {
	var projects []*model.Project
	response := client.Request(
		"GET",
		"api/v1/account/"+accountId+"/projects",
		nil,
	)

	data, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}

	logrus.WithField("payload", string(data)).Debug("Got project list response")

	err = json.Unmarshal(data, &projects)

	return projects, err
}
