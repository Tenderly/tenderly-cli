package call

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/model"
	"github.com/tenderly/tenderly-cli/rest/client"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type ProjectCalls struct {
}

func NewProjectCalls() *ProjectCalls {
	return &ProjectCalls{}
}

func (rest *ProjectCalls) CreateProject(request payloads.ProjectRequest) (*payloads.ProjectResponse, error) {
	projectJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var projectResponse payloads.ProjectResponse

	response := client.Request(
		"POST",
		"api/v1/account/"+config.GetString(config.AccountID)+"/project",
		projectJson,
	)
	err = json.NewDecoder(response).Decode(&projectResponse)

	if err != nil {
		return nil, fmt.Errorf("failed parsing create project respose: %s", err)
	}

	return &projectResponse, nil
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

func (rest *ProjectCalls) GetProjects(accountId string) (*payloads.GetProjectsResponse, error) {
	var getProjectsResponse payloads.GetProjectsResponse
	response := client.Request(
		"GET",
		"api/v1/account/"+accountId+"/projects?withShared=true",
		nil,
	)

	data, err := io.ReadAll(response)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &getProjectsResponse)

	if err != nil {
		return nil, fmt.Errorf("failed parsing get projects respose: %s", err)
	}

	for _, project := range getProjectsResponse.Projects {
		if string(project.Owner) != accountId {
			project.IsShared = true
		}
	}

	return &getProjectsResponse, nil
}
