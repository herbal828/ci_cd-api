package models

import (
	"time"
)

//PostRequestPayload represents the payload received in the POST request.
type PostRequestPayload struct {
	Repository struct {
		Name                *string  `json:"name"`
		RequireStatusChecks []string `json:"required_status_checks"`
	} `json:"repository"`

	Workflow struct {
		Type *string `json:"type"`
	} `json:"workflow"`

	CodeCoverage struct {
		PullRequestThreshold *float64 `json:"pull_request_threshold"`
	} `json:"code_coverage"`
}

//PutRequestPayload represents the payload received in the PUT request.
type PutRequestPayload struct {
	Fury struct {
		Technology *string `json:"technology"`
	} `json:"fury"`

	Repository struct {
		Name                *string
		RequireStatusChecks []string `json:"required_status_checks"`
	} `json:"repository"`

	CodeCoverage struct {
		PullRequestThreshold *float64 `json:"pull_request_threshold"`
	} `json:"code_coverage"`
}

//Configuration represents the only business object of this API.
//Has all the information needed for a good release process execution.
type Configuration struct {
	ID                               *string `gorm:"primary_key"`
	ApplicationName                  *string
	Technology                       *string
	RepositoryURL                    *string
	RepositoryStatusChecks           []RequireStatusCheck
	WorkflowType                     *string
	ContinuousIntegrationProvider    *string
	ContinuousIntegrationURL         *string
	BuildServerProvider              *string
	BuildServerURL                   *string
	CodeCoveragePullRequestThreshold *float64

	//GORM date attributes
	CreatedAt time.Time
	UpdatedAt time.Time
}

//RequireStatusCheck is a list of strings which represents all the status checks
//Required to be success before perform a 'git merge' in a protected branch.
type RequireStatusCheck struct {
	ID              *uint64 `gorm:"primary_key"`
	Check           string
	ConfigurationID *string
}

//NewConfiguration converts a PostRequestPayload into a Configuration.
func NewConfiguration(r *PostRequestPayload) *Configuration {
	var c Configuration

	c.ID = r.Repository.Name
	c.WorkflowType = r.Workflow.Type
	c.CodeCoveragePullRequestThreshold = r.CodeCoverage.PullRequestThreshold

	reqChecks := make([]RequireStatusCheck, 0)
	for _, rq := range r.Repository.RequireStatusChecks {
		reqChecks = append(reqChecks, RequireStatusCheck{
			Check: rq,
		})
	}

	c.RepositoryStatusChecks = reqChecks

	return &c
}

//UpdateConfiguration updates a Configuration based on a PutRequestPayload.
func (c *Configuration) UpdateConfiguration(r *PutRequestPayload) {
	if r.CodeCoverage.PullRequestThreshold != nil {
		c.CodeCoveragePullRequestThreshold = r.CodeCoverage.PullRequestThreshold
	}
	if r.Fury.Technology != nil {
		c.Technology = r.Fury.Technology
	}

	if r.Repository.RequireStatusChecks != nil {
		reqChecks := make([]RequireStatusCheck, 0)
		for _, rq := range r.Repository.RequireStatusChecks {
			reqChecks = append(reqChecks, RequireStatusCheck{
				Check: rq,
			})
		}
		c.RepositoryStatusChecks = reqChecks
	}
}

//GetRequiredStatusCheck maps the RepositoryStatusChecks field in the Configuration struct into a string slice.
func (c *Configuration) GetRequiredStatusCheck() []string {
	var rsc []string
	for _, rc := range c.RepositoryStatusChecks {
		rsc = append(rsc, rc.Check)
	}
	return rsc
}

//Marshall converts the Configuration struct into a readable JSON interface.
func (c *Configuration) Marshall() interface{} {
	rsc := c.GetRequiredStatusCheck()
	return &struct {
		ID   string `json:"id"`
		Fury struct {
			AppName string `json:"application_name"`
			Tech    string `json:"technology"`
		} `json:"fury"`
		Repository struct {
			URL                 string   `json:"url"`
			RequiredStatusCheck []string `json:"required_status_check"`
		} `json:"repository"`
		CI struct {
			Provider string `json:"provider"`
			URL      string `json:"url"`
		} `json:"continuous_integration"`
		BuildServer struct {
			Provider string `json:"provider"`
			URL      string `json:"url"`
		} `json:"build_server"`
		CodeCoverage struct {
			PullRequestThreshold float64 `json:"pull_request_threshold"`
		} `json:"code_coverage"`
		Workflow struct {
			Type string `json:"type"`
		} `json:"workflow"`
	}{
		*c.ID,
		struct {
			AppName string `json:"application_name"`
			Tech    string `json:"technology"`
		}{
			*c.ApplicationName,
			*c.Technology,
		},
		struct {
			URL                 string   `json:"url"`
			RequiredStatusCheck []string `json:"required_status_check"`
		}{
			*c.RepositoryURL,
			rsc,
		},
		struct {
			Provider string `json:"provider"`
			URL      string `json:"url"`
		}{
			*c.ContinuousIntegrationProvider,
			*c.ContinuousIntegrationURL,
		},
		struct {
			Provider string `json:"provider"`
			URL      string `json:"url"`
		}{
			*c.BuildServerProvider,
			*c.BuildServerURL,
		},
		struct {
			PullRequestThreshold float64 `json:"pull_request_threshold"`
		}{
			*c.CodeCoveragePullRequestThreshold,
		},
		struct {
			Type string `json:"type"`
		}{
			*c.WorkflowType,
		},
	}
}


type PostPerformanceBody struct {
	ApplicationName string `json:"application_name"`
	Start           string `json:"start"`
	End             string `json:"end"`
	Scope           string `json:"scope"`
	RollupUnit      string `json:"rollup-unit"`
}
