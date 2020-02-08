package services

import (
	"github.com/herbal828/ci_cd-api/api/models"
)
import "github.com/herbal828/ci_cd-api/api/configs"

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type WorkflowService interface {
	SetWorkflow(config *models.WorkflowConfig) error
}

//SetWorkflow protects the necessary branches for the workflow selected by the user
//It performs all the actions needed to enabled successfuly Release Process.
func (c Configuration) SetWorkflow(config *models.Configuration) error {

	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	for _, branch := range workflowBranchesList {
		if branch.Stable {
			bpError := c.GithubClient.ProtectBranch(config, &branch)

			if bpError != nil {
				//the branch does not exist. We will create it.
				if bpError.Error() == "branch not found" {
					//TODO: Si el branch NO existe. Crearlo, asi nos ahorramos hacer un get todo el tiempo
				}
			}
		}
	}

	return nil
}
