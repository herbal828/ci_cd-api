package services

import (
	"errors"
	"fmt"
	"github.com/herbal828/ci_cd-api/src/api/clients"
	"github.com/herbal828/ci_cd-api/src/api/models"
	"github.com/herbal828/ci_cd-api/src/api/services/storage"
	"github.com/herbal828/ci_cd-api/src/api/utils"
	"github.com/jinzhu/gorm"
	"os"
	"time"


)

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type ConfigurationService interface {
	Create(*models.PostRequestPayload) (*models.Configuration, error)
	Get(string) (*models.Configuration, error)
	Update(r *models.PutRequestPayload) (*models.Configuration, error)
	Delete(id string) error
	GetInserts() error
	GetAppPerformance(req *models.PostPerformanceBody) error
}

//Configuration represents the ConfigurationService layer
//It has an instance of a DBClient layer and
//A Logger to perform all the log actions.
type Configuration struct {
	SQL            storage.SQLStorage
	Logger         logs.Logger
	FuryClient     clients.FuryClient
	BuilderClient  clients.BuilderClient
	MelicovClient  clients.MelicovClient
	WorkflowClient clients.WorkflowClient
	NetRPClient    clients.NetRPClient
	MantraClient   clients.MantraClient
}

//NewConfigurationService initializes a ConfigurationService
func NewConfigurationService(sql storage.SQLStorage) *Configuration {
	return &Configuration{
		SQL: sql,
		Logger: &logs.Log{
			Component: "ConfigurationService",
		},
		FuryClient:     clients.NewFuryClient(),
		BuilderClient:  clients.NewBuilderClient(),
		MelicovClient:  clients.NewMelicovClient(),
		WorkflowClient: clients.NewWorkflowClient(),
		NetRPClient:    clients.NewNetRPClient(),
		MantraClient:   clients.NewMantraClient(),
	}
}

//Create creates a Release Process valid configuration.
//It performs all the actions needed to enabled successfuly Release Process.
func (s *Configuration) Create(r *models.PostRequestPayload) (*models.Configuration, error) {

	config := *models.NewConfiguration(r)

	repoName := config.ID
	var cf models.Configuration

	//Search the configuration into database
	if err := s.SQL.GetBy(&cf, "id = ?", *repoName); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, errors.New("error checking configuration existence")
		}

		//If the configuration doesn't exist, then create it

		//Set Fury values like application name, repo url & technology.
		if err := s.FuryClient.GetApplicationData(&config); err != nil {
			return nil, err
		}

		//Set Workflow
		if err := s.WorkflowClient.SetWorkflow(&config); err != nil {
			return nil, err
		}

		//Create Continuous Integration and Build Server
		if err := s.BuilderClient.CreateJob(&config); err != nil {
			//If something was wrong, then rollback workflow
			if err := s.WorkflowClient.UnSetWorkflow(&config); err != nil {
				return nil, err
			}
			return nil, err
		}

		if err := s.MelicovClient.SetPullRequestThreshold(&config); err != nil {
			//If something was wrong, then do nothing by now TODO: Change it in melicov
		}

		//Enable app in Net RP Api
		if err := s.NetRPClient.EnableReleaseProcess(&config); err != nil {
			//If something was wrong, then rollback workflow ...
			if err := s.WorkflowClient.UnSetWorkflow(&config); err != nil {
				return nil, err
			}
			//... and CI & Build Server
			if err := s.BuilderClient.DeleteJob(&config); err != nil {
				return nil, err
			}
			return nil, err
		}

		//Enable app in Fury API
		if err := s.FuryClient.EnableReleaseProcessField(&config); err != nil {
			//If something was wrong, then rollback workflow ...
			if err := s.WorkflowClient.UnSetWorkflow(&config); err != nil {
				return nil, err
			}
			//...CI & Build Server
			if err := s.BuilderClient.DeleteJob(&config); err != nil {
				return nil, err
			}

			//... and Net RP
			if err := s.NetRPClient.DisableReleaseProcess(&config); err != nil {
				return nil, err
			}

			return nil, err
		}

		//Save it into database
		if err := s.SQL.Insert(&config); err != nil {
			return nil, errors.New("error saving new configuration")
		}
		return &config, nil

	} else { //If configuration already exists then return it
		return &cf, nil
	}
}

//Get searches a configuration into database.
//Returns an error if the config is not found.
func (s *Configuration) Get(id string) (*models.Configuration, error) {
	var cf models.Configuration
	if err := s.SQL.GetBy(&cf, "id = ?", id); err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, errors.New("error checking configuration existance")
		}
		return nil, err
	}
	return &cf, nil
}

//Update modifies a configuration.
//It receives a PutRequestPayload.
//Returns an error if the config is not found or if it some problem updating the config.
func (s *Configuration) Update(r *models.PutRequestPayload) (*models.Configuration, error) {

	oldConfig, err := s.Get(*r.Repository.Name)

	if err != nil {
		return nil, err
	}

	newConfig := *oldConfig
	newConfig.UpdateConfiguration(r)

	//Update the repository status checks
	if r.Repository.RequireStatusChecks != nil {
		if err := s.WorkflowClient.SetWorkflow(&newConfig); err != nil {
			return nil, err
		}

		//TODO: Change this, because it is a change made in order to be able to update the required status checks
		//we did this because when we updated the fields, it doesn't update them in the require_status_check
		// child table, so we removed them and then saved the new ones.

		//Delete from configurations DB
		if sqlErr := s.SQL.DeleteFromRequireStatusChecksByConfigurationID(oldConfig.ID); sqlErr != nil {
			return nil, sqlErr
		}

	}

	//Update the repository technology
	if r.Fury.Technology != nil {
		//Update CI Job
		if err := s.BuilderClient.CreateJob(&newConfig); err != nil {

			//if workflow was updated we need to rollback
			if r.Repository.RequireStatusChecks != nil {
				if err := s.WorkflowClient.UnSetWorkflow(oldConfig); err != nil {
					return nil, err
				}
			}
			return nil, err
		}

		//Update fury application technology
		if err := s.FuryClient.UpdateApplicationTechnology(&newConfig); err != nil {

			//if workflow was updated we need to rollback workflow and CI pipeline
			if r.Repository.RequireStatusChecks != nil {
				//Rollback Workflow
				setWorkflowErr := s.WorkflowClient.SetWorkflow(oldConfig)
				createJobErr := s.BuilderClient.CreateJob(oldConfig)

				if setWorkflowErr != nil || createJobErr != nil {
					return nil, err
				}
			}
			return nil, err
		}
	}

	if r.CodeCoverage.PullRequestThreshold != nil {
		if err := s.MelicovClient.UpdatePullRequestThreshold(&newConfig); err != nil {
			//If something was wrong, then do nothing by now TODO: Change it in melicov
		}
	}

	//Save the new config into database
	if err := s.SQL.Update(&newConfig); err != nil {
		return nil, errors.New("error updating repository configuration")
	}
	return &newConfig, nil
}

//Delete erase the configuration.
//It makes a sof delete.
//Receives the configuration id (repoName) and returns an error it it occurs.
func (s *Configuration) Delete(id string) error {

	cf, err := s.Get(id)

	if err != nil {
		return err
	}

	//Unset Workflow
	if unsetErr := s.WorkflowClient.UnSetWorkflow(cf); unsetErr != nil {
		return unsetErr
	}

	//Delete CI Job
	if ciErr := s.BuilderClient.DeleteJob(cf); ciErr != nil {
		return ciErr
	}

	//Remove repository from netRP
	if netErr := s.NetRPClient.DisableReleaseProcess(cf); netErr != nil {
		return netErr
	}

	//Turn OFF rp on fury
	if furyErr := s.FuryClient.DisableReleaseProcessField(cf); furyErr != nil {
		return furyErr
	}

	//Delete from configurations DB
	if sqlErr := s.SQL.Delete(cf); sqlErr != nil {
		return sqlErr
	}

	return nil
}

func (s *Configuration) GetInserts() error {

	//Creo los 2 files
	configurationFile, err1 := os.Create("/tmp/config_inserts.txt")
	statusFile, err2 := os.Create("/tmp/status_inserts.txt")

	defer configurationFile.Close()
	defer statusFile.Close()

	if err1 != nil {
		return errors.New("No se pudo crear el archivo configs_insert")
	}

	if err2 != nil {
		return errors.New("No se pudo crear el archivo status_inserts")
	}

	//var insertSlice []string

	rpList, error := s.NetRPClient.GetReleaseProcessEnabledApplicationList()

	if error != nil {
		return errors.New("Error al traer la lista de apps en Net RP")
	}

	cleanList := removeDuplicates(rpList)

	for i, repo := range cleanList {

		if i == 3 {
			break
		}
		var config models.Configuration
		config.ID = utils.Stringify(repo)

		if err := s.FuryClient.GetApplicationData(&config); err != nil {
			fmt.Printf("Error getting fury data for: %s", config.ID)
		}

		config.WorkflowType = utils.Stringify("gitflow")
		config.ContinuousIntegrationURL = utils.Stringify(fmt.Sprintf("https://rp-ci.furycloud.io/job/%s/", *config.ApplicationName))
		config.ContinuousIntegrationProvider = utils.Stringify("JENKINS")
		config.BuildServerURL = utils.Stringify(fmt.Sprintf("https://rp-ci.furycloud.io/job/%s/", *config.ApplicationName))
		config.BuildServerProvider = utils.Stringify("JENKINS")

		configInsertString := "INSERT INTO `configurations` (id,application_name,technology,repository_url, workflow_type,continuous_integration_provider,continuous_integration_url,build_server_provider,build_server_url,code_coverage_pull_request_threshold,created_at, updated_at) VALUES('%s', '%s', '%s', '%s' ,'%s', '%s', '%s', '%s', '%s', %d, '%s', '%s');"

		var configStringToAdd string
		configStringToAdd = fmt.Sprintf(configInsertString, *config.ID, *config.ApplicationName, *config.Technology, *config.RepositoryURL, *config.WorkflowType, *config.ContinuousIntegrationProvider, *config.ContinuousIntegrationURL, *config.BuildServerProvider, *config.BuildServerURL, 80, time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

		//Lo agrego al archivo
		fmt.Fprintln(configurationFile, configStringToAdd)

		var statusStringToAdd string

		statusChecksList := []string{"continuous-integration", "pull-request-coverage", "workflow"}

		//Por cada status agregamos un insert
		for _, status := range statusChecksList {
			StatusChecksInsertString := "INSERT INTO `require_status_checks` (`check`, configuration_id) VALUES('%s','%s');"
			statusStringToAdd = fmt.Sprintf(StatusChecksInsertString, status, *config.ID)
			fmt.Fprintln(statusFile, statusStringToAdd)
		}
	}

	return nil
}

func removeDuplicates(elements []string) []string { // change string to int here if required
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{} // change string to int here if required
	result := []string{}             // change string to int here if required

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func (s *Configuration) GetAppPerformance(req *models.PostPerformanceBody) error {

	if err := s.MantraClient.GetApplicationPerformance(req); err != nil {
		return err
	}

	return nil

}
