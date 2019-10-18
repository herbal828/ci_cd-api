package controllers

import (
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
)

//HTTPContext defines all the
type HTTPContext interface {
	BindJSON(interface{}) error
	GetHeader(string) string
	JSON(int, interface{})
	Param(key string) string
}

//Configuration represents the ConfigurationController layer
//It has an instance of a ConfigurationService layer and
//A Logger to perform all the log actions.
type Configuration struct {
	Service services.ConfigurationService
	Logger  logs.Logger
}

//NewConfigurationController initializes a ConfigurationController
func NewConfigurationController(sql storage.SQLStorage) *Configuration {
	return &Configuration{
		Service: services.NewConfigurationService(sql),
		Logger: &logs.Log{
			Component: "ConfigurationController",
		},
	}
}

//Create creates a new configuration for the given repository
//It could returns
//	200OK in case of a success processing the creation
//	400BadRequest in case of an error parsing the request payload
//	500InternalServerError in case of an internal error procesing the creation
func (c *Configuration) Create(ctx HTTPContext) {
	var req models.PostRequestPayload
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			apierrors.NewBadRequestApiError("invalid configuration request payload"),
		)
		return
	}

	config, err := c.Service.Create(&req)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			apierrors.NewInternalServerApiError("something was wrong creating a new configuration", err),
		)
		return
	}

	ctx.JSON(http.StatusOK, config.Marshall())
}

//Show retrieves the configuration for a given repository.
//It could returns
//	200OK in case of a success procesing the search
//	404NotFound in case of the non existance of the configuration
//	500InternalServerError in case of an internal error procesing the search
func (c *Configuration) Show(ctx HTTPContext) {
	repoName := getRepoNamefromURL(ctx)
	config, err := c.Service.Get(repoName)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			ctx.JSON(
				http.StatusInternalServerError,
				apierrors.NewInternalServerApiError(fmt.Sprintf("something was wrong getting the configuration for %s", repoName), err),
			)
			return
		}
		ctx.JSON(
			http.StatusNotFound,
			apierrors.NewNotFoundApiError(fmt.Sprintf("configuration for repository %s not found", repoName)),
		)
		return
	}

	ctx.JSON(http.StatusOK, config.Marshall())
}

//Update updates the configuration for a given repository.
//It could returns
//	200OK in case of a success procesing the update
//	404NotFound in case of the non existance of the configuration
//	500InternalServerError in case of an internal error procesing the search
func (c *Configuration) Update(ctx HTTPContext) {
	var req models.PutRequestPayload
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			apierrors.NewBadRequestApiError("invalid configuration request payload"),
		)
		return
	}

	repoName := getRepoNamefromURL(ctx)
	req.Repository.Name = &repoName

	config, err := c.Service.Update(&req)

	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			apierrors.NewInternalServerApiError("something was wrong updating repository configuration", err),
		)
	}

	ctx.JSON(http.StatusOK, config.Marshall())

}

//Delete erases the configuration for a given repository from db, turn off Workflow and deletes continuous integration Jobs.
//It could returns
//	204NoContent in case of a success procesing the delete
//	404NotFound in case of the non existance of the configuration
//	500InternalServerError in case of an internal error processing the delete
func (c *Configuration) Delete(ctx HTTPContext) {

	repoName := getRepoNamefromURL(ctx)
	err := c.Service.Delete(repoName)

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			ctx.JSON(
				http.StatusInternalServerError,
				apierrors.NewInternalServerApiError(fmt.Sprintf("something was wrong getting the configuration for %s", repoName), err),
			)
			return
		}
		ctx.JSON(
			http.StatusNotFound,
			apierrors.NewNotFoundApiError(fmt.Sprintf("configuration for repository %s not found", repoName)),
		)
		return
	}

	ctx.JSON(
		http.StatusNoContent,
		nil,
	)
}

func getRepoNamefromURL(ctx HTTPContext) string {
	return ctx.Param("repoName")
}

func (c *Configuration) GetDBInserts(ctx HTTPContext) {

	error := c.Service.GetInserts()
	if error != nil {
		fmt.Println("Error obteniendo la lista")
	}
	ctx.JSON(http.StatusOK, error)
}

func (c *Configuration) GetAppPerformance(ctx HTTPContext) {

	var req models.PostPerformanceBody
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			apierrors.NewBadRequestApiError("invalid performance body"),
		)
		return
	}

	error := c.Service.GetAppPerformance(&req)

	if error != nil {
		fmt.Println("Error obteniendo la performance")
	}

	ctx.JSON(http.StatusOK, error)
}
