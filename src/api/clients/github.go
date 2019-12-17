package clients

// Builder client, connects configurations API with the CI proxy API
// and implements the necessary functions to
// create and delete jobs necessary for the execution of release process

import (
	"errors"
	"fmt"
	"github.com/herbal828/ci_cd-api/src/api/configs"
	"github.com/herbal828/ci_cd-api/src/api/models"
	"github.com/herbal828/ci_cd-api/src/api/utils"
	"github.com/mercadolibre/golang-restclient/rest"
	"net/http"
	"time"
)

type GithubClient interface {
	GetBranch(config *models.Configuration, branchName string) error
}

type githubClient struct {
	Client Client
}

func NewGithubClient() GithubClient {
	hs := make(http.Header)
	hs.Set("cache-control", "no-cache")
	hs.Set("Authorization", "token c418a866d0dad4374b231fbc9726bb0407ada038")
	hs.Set("Accept", "application/vnd.github.luke-cage-preview+json")

	return &githubClient{
		Client: &client{
			RestClient: &rest.RequestBuilder{
				BaseURL:        configs.GetGithubBaseURL(),
				Timeout:        2 * time.Second,
				Headers:        hs,
				ContentType:    rest.JSON,
				DisableCache:   true,
				DisableTimeout: false,
			},
		},
	}
}

type ghGetBranchResponse struct {
	Message  string `json:"message"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

//Gets a repository branch info
//This perform a GET request to Github api using
func (c *githubClient) GetBranch(config *models.Configuration, branchName string) error {

	if config.RepositoryName == nil || config.RepositoryOwner == nil || branchName == "" {
		err := errors.New("invalid github body params")
		return err
	}

	response := c.Client.Get(fmt.Sprintf("/repos/%s/%s/branches/%s", *config.RepositoryOwner, *config.RepositoryName, branchName))

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK {
		return errors.New("error getting repository branch")
	}

	return nil
}

//SetWorkflow enables the workflow selected in Github
//Protects the stable branches with the given required status checks
//It returns an error if it occurs.
func (c *githubClient) SetWorkflow(config *models.Configuration) error {

	//Check if the needed configuration values are ok
	if config.WorkflowType == nil || config.ID == nil {
		err := errors.New("invalid workflow body params")
		return err
	}

	switch *config.WorkflowType {
	case "gitflow":


	}

	//TODO: Hacer el switch por workflow
	//TODO: solo vamos a aceptar workflow gitflow
	//TODO: Traer los branches a proteger.
	//TODO: recorrer cada branch, ver si existe. Si no existe, crear el branch y luego protegerlo.




	body := map[string]interface{}{
		"type":                   *config.WorkflowType,
		"repository_name":        *config.ID,
		"required_status_checks": config.GetRequiredStatusCheck(),
	}

	response := c.Client.Post("/workflow", body)

	if response.Err() != nil {

		return response.Err()
	}

	if utils.WasOK(response.StatusCode()) && utils.WasCreated(response.StatusCode()) {
		return errors.New("error setting workflow")
	}

	return nil
}
