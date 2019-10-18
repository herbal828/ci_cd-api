package clients

// Builder client, connects configurations API with the CI proxy API
// and implements the necessary functions to
// create and delete jobs necessary for the execution of release process

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/herbal828/ci_cd-api/src/api/configs"
	"github.com/herbal828/ci_cd-api/src/api/models"
	"net/http"
	"time"
	 "github.com/mercadolibre/golang-restclient/rest"
)

type BuilderClient interface {
	CreateJob(config *models.Configuration) error
	DeleteJob(config *models.Configuration) error
}

type builderClient struct {
	Client Client
}

func NewBuilderClient() BuilderClient {
	hs := make(http.Header)
	hs.Set("Content-Type", "application/json")

	return &builderClient{
		Client: &client{
			RestClient: &rest.RequestBuilder{
				BaseURL:        configs.GetCIProxyBaseURL(),
				Timeout:        2 * time.Second,
				Headers:        hs,
				ContentType:    rest.JSON,
				DisableCache:    true,
				DisableTimeout: false,
			},
		},
	}
}

type builderResponse struct {
	Message  string `json:"message"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

//Creates a job in ci and builder
//
//
//This perform a Post request to CI-Proxy api with following body params:
//name: fury application Name
//technology: application technology f.e ("java", "go" ..)
//repository_url: github repository url
func (c *builderClient) CreateJob(config *models.Configuration) error {

	if config.ID == nil || config.ApplicationName == nil || config.Technology == nil || config.RepositoryURL == nil {
		err := errors.New("invalid ci-proxy body params")
		return err
	}

	body := map[string]interface{}{
		"name":           *config.ApplicationName,
		"technology":     *config.Technology,
		"repository_url": *config.RepositoryURL,
	}

	response := c.Client.Post("/job", body)

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {


		return errors.New("error creating job")
	}

	var info builderResponse
	if err := json.Unmarshal(response.Bytes(), &info); err != nil {

		return errors.New("error binding ci-proxy response")
	}

	config.ContinuousIntegrationProvider = &info.Provider
	config.ContinuousIntegrationURL = &info.URL

	config.BuildServerProvider = &info.Provider
	config.BuildServerURL = &info.URL



	return nil
}

//Delete ci & builder jobs from CI-Proxy API
//
//This perform a DELETE request to CI-Proxy api using the fury application name
func (c *builderClient) DeleteJob(config *models.Configuration) error {

	if config.ApplicationName == nil {
		err := errors.New("invalid application name param")
		return err
	}



	response := c.Client.Delete(fmt.Sprintf("/job/%s", *config.ApplicationName))

	if response.Err() != nil {

		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusNoContent {
		return errors.New("error deleting job")
	}

	return nil
}
