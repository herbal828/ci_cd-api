package configs

import "github.com/herbal828/ci_cd-api/api/models"

func GetWorkflowConfiguration(configuration *models.Configuration) *models.WorkflowConfig {
	var workflowConfig models.WorkflowConfig

	switch *configuration.WorkflowType {
	case "gitflow":
		workflowConfig = *GetGitflowConfig(configuration)
	default:
		workflowConfig = *GetGitflowConfig(configuration)
	}

	return &workflowConfig
}

func GetGitflowConfig(configuration *models.Configuration) *models.WorkflowConfig {

	var masterRequirements models.Requirements
	var masterWorkflowRequiredStatusChecks models.RequiredStatusChecks

	//Branch Master

	masterWorkflowRequiredStatusChecks.IncludeAdmins = true
	masterWorkflowRequiredStatusChecks.Strict = true
	masterWorkflowRequiredStatusChecks.Contexts = GetRequiredStatusCheck(configuration)

	masterRequirements.EnforceAdmins = true
	masterRequirements.AcceptPrFrom = []string{"release", "hotfix"}
	masterRequirements.RequiredStatusChecks = masterWorkflowRequiredStatusChecks

	masterBranchConfig := models.Branch{
		Requirements: masterRequirements,
		Stable:       true,
		Name:         "master",
		Releasable:   true,
		StartWith:    false,
	}

	//Develop Branch

	var developRequirements models.Requirements
	var developWorkflowRequiredStatusChecks models.RequiredStatusChecks

	developRequirements.EnforceAdmins = true
	developRequirements.AcceptPrFrom = []string{"feature", "fix", "enhancement", "bugfix"}

	developWorkflowRequiredStatusChecks.IncludeAdmins = true
	developWorkflowRequiredStatusChecks.Strict = true
	developWorkflowRequiredStatusChecks.Contexts = GetRequiredStatusCheck(configuration)

	developBranchConfig := models.Branch{
		Requirements: developRequirements,
		Stable:       true,
		Name:         "develop",
		Releasable:   false,
		StartWith:    false,
	}

	//Build the gitflow configuration

	gfConfig := models.WorkflowConfig{
		Name: "gitflow",
		Description: models.Description{
			Branches: []models.Branch{
				masterBranchConfig,
				developBranchConfig,
			},
		},
		Detail: "Workflow Description",
	}

	return &gfConfig
}

//GetRequiredStatusCheck maps the RepositoryStatusChecks field in the Configuration struct into a string slice.
func GetRequiredStatusCheck(c *models.Configuration) []string {
	var rsc []string
	for _, rc := range c.RepositoryStatusChecks {
		rsc = append(rsc, rc.Check)
	}
	return rsc
}
