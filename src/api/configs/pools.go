package configs

import "os"

const (
	workflowProductionBaseURL = "http://production.rp-workflow-api.melifrontends.com"
	ciProxyProductionBaseURL  = "http://production.rp-ci-proxy.melifrontends.com"
	melicovProductionBaseURL  = "http://production.melicov-server.melifrontends.com"
	netrpProductionBaseURL    = "http://release-process.netrp-proxy.melifrontends.com"
	furyProductionBaseURL     = "http://api.furycloud.io"
	mantraProductionBaseURL   = "http://api.mp.internal.ml.com"
)

const (
	workflowStageBaseURL = "http://test.rp-workflow-api.melifrontends.com"
	ciProxyStageBaseURL  = "http://test.rp-ci-proxy.melifrontends.com"
	melicovStageBaseURL  = "http://test.melicov-server.melifrontends.com"
	netrpStageBaseURL    = "http://release-process.netrp-proxy.melifrontends.com"
	furyStageBaseURL     = "http://fury-api.melifrontends.com"
)

const (
	workflowLocalBaseURL = "http://localhost:7777"
	ciProxyLocalBaseURL  = "http://localhost:8888"
	melicovLocalBaseURL  = "http://localhost:9999"
	netrpLocalBaseURL    = "http://localhost:8090"
	furyLocalBaseURL     = "http://localhost:5555"
)

//GetWorkflowBaseURL returns the workflow API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetWorkflowBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return workflowProductionBaseURL
	case "test":
		return workflowStageBaseURL
	default:
		return workflowLocalBaseURL
	}
}

//GetCIProxyBaseURL returns the ci-proxy API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetCIProxyBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return ciProxyProductionBaseURL
	case "test":
		return ciProxyStageBaseURL
	default:
		return ciProxyLocalBaseURL
	}
}

//GetMelicovBaseURL returns the melicov API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetMelicovBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return melicovProductionBaseURL
	case "test":
		return melicovStageBaseURL
	default:
		return melicovLocalBaseURL
	}
}

//GetNetRPBaseURL returns the netrp API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetNetRPBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return netrpProductionBaseURL
	case "test":
		return netrpProductionBaseURL
	default:
		return netrpProductionBaseURL
	}
}

//GetFuryBaseURL returns the fury API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetFuryBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return furyProductionBaseURL
	case "test":
		return furyProductionBaseURL
	default:
		return furyProductionBaseURL
	}
}

//GetMantraBaseURL returns the ci-proxy API url based on the current environment.
//Valid environments are: 'production' for production,
//'test' for a stage scope and
//any other will be considered as a local environment.
func GetMantraBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return mantraProductionBaseURL
	case "test":
		return mantraProductionBaseURL
	default:
		return mantraProductionBaseURL
	}
}
