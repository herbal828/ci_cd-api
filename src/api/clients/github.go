package clients

// Builder client, connects configurations API with the CI proxy API
// and implements the necessary functions to
// create and delete jobs necessary for the execution of release process

import (
	"errors"
	"fmt"
	"github.com/herbal828/ci_cd-api/src/api/configs"
	"github.com/herbal828/ci_cd-api/src/api/models"
	"github.com/mercadolibre/golang-restclient/rest"
	"net/http"
	"time"
)

type GithubClient interface {
	GetBranch(config *models.Configuration, branchName string) error
	ProtectBranch(config *models.Configuration, branchConfig *models.Branch) error
}

type githubClient struct {
	Client Client
}

func NewGithubClient() GithubClient {
	hs := make(http.Header)
	hs.Set("cache-control", "no-cache")
	hs.Set("Authorization", "token 051fd6be26af16cd7e8f3c9ecd54c18845f6b074")
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

func (c *githubClient) ProtectBranch(config *models.Configuration, branchConfig *models.Branch) error {

	if branchConfig.Name == "" {
		err := errors.New("invalid branch protection body params")
		return err
	}

	body := map[string]interface{}{
		"enforce_admins":                true,
		"required_status_checks":        branchConfig.Requirements.RequiredStatusChecks,
		"required_pull_request_reviews": branchConfig.Requirements.RequiredPullRequestReviews,
	}

	response := c.Client.Put(fmt.Sprintf("/%p/branches/%s/protection", &config.RepositoryName, branchConfig.Name), body)

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		if response.StatusCode() == http.StatusNotFound {
			return errors.New("branch not found")
		}
		return errors.New(fmt.Sprintf("error protecting branch - status: %d", response.StatusCode()))
	}

	return nil
}
