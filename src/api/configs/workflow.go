package configs

import "github.com/herbal828/ci_cd-api/src/api/models"

func getGitflowConfig() models.WorkflowConfig {
	var gfConfig models.WorkflowConfig
	var gfDescription models.Description

	var masterGFConfig models.Branch
	var developGFConfig models.Branch
	var releaseGFConfig models.Branch

	masterGFConfig.Name = "master"
	masterGFConfig.Releasable = true
	masterGFConfig.Stable = true
	masterGFConfig.StartWith = false

	var masterRequiriments models.Requirements
	masterRequiriments.EnforceAdmins = true

	masterRequiriments.AcceptPrFrom = []string{"release", "hotfix"}



	gfConfig.Name = "gitflow"
	gfConfig.Detail = "Breve descripcion del workflow"



	//gfRequirements.AcceptPrFrom = []string{"Penn", "Teller"}


	return gfConfig
}

func GetWorkflowConfig(workflow string) string {

}
