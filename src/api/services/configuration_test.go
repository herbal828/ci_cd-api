package services

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	"github.com/mercadolibre/golang-restclient/rest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

)

func TestMain(m *testing.M) {
	rest.StartMockupServer()
	os.Exit(m.Run())
}

func TestConfiguration_Create(t *testing.T) {

	type workflowFuncResult struct {
		setWorkflow   error
		unsetWorkflow error
	}

	type builderFuncResult struct {
		createJob error
		deleteJob error
	}

	type clientsResult struct {
		furyClient     error
		workflowClient workflowFuncResult
		builderClient  builderFuncResult
		netRPClient    error
		melicovClient  error
		sqlClient      error
	}

	type workflowClientFuncExecution struct {
		execute       bool
		setWorkflow   bool
		unsetWorkflow bool
	}

	type builderClientFuncExecution struct {
		execute   bool
		createJob bool
		deleteJob bool
	}

	type clientsExecution struct {
		furyClient     bool
		workflowClient workflowClientFuncExecution
		builderClient  builderClientFuncExecution
		netRPClient    bool
		melicovClient  bool
	}

	type args struct {
		postRequestPayload *models.PostRequestPayload
		config             *models.Configuration
		execute            clientsExecution
	}

	type expects struct {
		errorLog      string
		infoLog       string
		error         error
		config        models.Configuration
		clientsResult clientsResult
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.Workflow.Type = utils.Stringify("gitflow")
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	tests := []struct {
		name          string
		args          args
		wantErr       bool
		expects       expects
		clientsResult clientsResult
	}{
		{
			name: "test Create - configuration already exists then return it",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test Create - Error checking configuration existence, should return an Error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
			},
			expects: expects{
				error: errors.New("record not found"),
			},
			wantErr: true,
		},
		{
			name: "test Create - Error checking configuration existence, should return an Error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
			},
			expects: expects{
				error: errors.New("some error"),
			},
			wantErr: true,
		},
		{
			name: "test Create - Fails to get fury application information, we return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					furyClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - Fails to post workflow config, we return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:     true,
						setWorkflow: true,
					},
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					workflowClient: workflowFuncResult{
						setWorkflow: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - Fails to post ci & builder Job, should unset workflow with error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:       true,
						setWorkflow:   true,
						unsetWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
					},
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					workflowClient: workflowFuncResult{
						unsetWorkflow: errors.New("some error"),
					},
					builderClient: builderFuncResult{
						createJob: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - enable rp on netrp and unsetworkflow fails, should return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:       true,
						setWorkflow:   true,
						unsetWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					netRPClient: errors.New("some error"),
					workflowClient: workflowFuncResult{
						unsetWorkflow: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - enable rp on netrp and delete ci job fails, should return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:       true,
						setWorkflow:   true,
						unsetWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
						deleteJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					netRPClient: errors.New("some error"),
					builderClient: builderFuncResult{
						deleteJob: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - enable rp on netrp, should rollback config without errors",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:       true,
						setWorkflow:   true,
						unsetWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
						deleteJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					netRPClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - Error saving on DB, should return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:     true,
						setWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					sqlClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Create - configuration created successfully",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:     true,
						setWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
			},
			wantErr: false,
		},
		{
			name: "test Create - enable rp on netrp and delete ci job fails, should return an error",
			args: args{
				postRequestPayload: &payload,
				config:             &repoConfigOK,
				execute: clientsExecution{
					furyClient: true,
					workflowClient: workflowClientFuncExecution{
						execute:       true,
						setWorkflow:   true,
						unsetWorkflow: true,
					},
					builderClient: builderClientFuncExecution{
						execute:   true,
						createJob: true,
						deleteJob: true,
					},
					melicovClient: true,
					netRPClient:   true,
				},
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				clientsResult: clientsResult{
					netRPClient: errors.New("some error"),
					builderClient: builderFuncResult{
						deleteJob: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Release Process Clients
			furyClient := interfaces.NewMockFuryClient(ctl)
			builderClient := interfaces.NewMockBuilderClient(ctl)
			workflowClient := interfaces.NewMockWorkflowClient(ctl)
			melicovClient := interfaces.NewMockMelicovClient(ctl)
			netRPClient := interfaces.NewMockNetRPClient(ctl)

			//SQL Clients
			sqlStorage := interfaces.NewMockSQLStorage(ctl)

			//Logger
			logger := interfaces.NewMockLogger(ctl)

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
					return &tt.expects.config
				}).
				Return(tt.expects.error).
				AnyTimes()

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			//Fury Client
			if tt.args.execute.furyClient {
				furyClient.EXPECT().
					GetApplicationData(gomock.Any()).
					Return(tt.expects.clientsResult.furyClient).
					AnyTimes()

				furyClient.EXPECT().
					EnableReleaseProcessField(gomock.Any()).
					Return(tt.expects.clientsResult.furyClient).
					AnyTimes()
			}

			//Workflow Client
			if tt.args.execute.workflowClient.execute {
				//SetWorkflow
				if tt.args.execute.workflowClient.setWorkflow {
					workflowClient.EXPECT().
						SetWorkflow(gomock.Any()).
						Return(tt.expects.clientsResult.workflowClient.setWorkflow).
						AnyTimes()
				}
				//UnsetWorkflow
				if tt.args.execute.workflowClient.unsetWorkflow {
					workflowClient.EXPECT().
						UnSetWorkflow(gomock.Any()).
						Return(tt.expects.clientsResult.workflowClient.unsetWorkflow).
						AnyTimes()
				}
			}

			//Builder Client
			if tt.args.execute.builderClient.execute {
				//Create Job
				if tt.args.execute.builderClient.createJob {
					builderClient.EXPECT().
						CreateJob(gomock.Any()).
						Return(tt.expects.clientsResult.builderClient.createJob).
						AnyTimes()
				}
				//Delete Job
				if tt.args.execute.builderClient.deleteJob {
					builderClient.EXPECT().
						DeleteJob(gomock.Any()).
						Return(tt.expects.clientsResult.builderClient.deleteJob).
						AnyTimes()
				}
			}

			//Melicov Client
			if tt.args.execute.melicovClient {
				melicovClient.EXPECT().
					SetPullRequestThreshold(gomock.Any()).
					Return(tt.expects.clientsResult.melicovClient).
					AnyTimes()
			}

			//NetRP Client
			if tt.args.execute.netRPClient {
				netRPClient.EXPECT().
					EnableReleaseProcess(gomock.Any()).
					Return(tt.expects.clientsResult.netRPClient).
					AnyTimes()
			}

			s := &Configuration{
				SQL:            sqlStorage,
				Logger:         logger,
				FuryClient:     furyClient,
				WorkflowClient: workflowClient,
				BuilderClient:  builderClient,
				MelicovClient:  melicovClient,
				NetRPClient:    netRPClient,
			}

			_, err := s.Create(tt.args.postRequestPayload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestConfiguration_Create_full(t *testing.T) {

	type expects struct {
		errorLog string
		infoLog  string
		error    error
		config   *models.Configuration
	}

	type args struct {
		postRequestPayload *models.PostRequestPayload
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.Workflow.Type = utils.Stringify("gitflow")
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test Create - Todo bien",
			args: args{
				postRequestPayload: &payload,
			},
			expects: expects{
				config: &repoConfigOK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Configuration Service
			service := interfaces.NewMockConfigurationService(ctl)

			service.
				EXPECT().
				Create(gomock.Any()).
				Return(tt.expects.config, tt.expects.error).
				AnyTimes()

			got, err := service.Create(tt.args.postRequestPayload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != nil {
				assert.Equal(t, "fury_repo-name", *got.ID, "the IDs should be equals")
				assert.Equal(t, "repo-name", *got.ApplicationName, "the ApplicationName should be equals")
				assert.Equal(t, "java", *got.Technology, "the Technology should be equals")
				assert.Equal(t, "http://github.com/fury_repo-name", *got.RepositoryURL, "the RepositoryURL should be equals")
				assert.Equal(t, "https://rp-ci.furycloud.io/job/repo-name/", *got.ContinuousIntegrationURL, "the ContinuousIntegrationURL should be equals")
				assert.Equal(t, "https://rp-builds.furycloud.io/job/repo-name/", *got.BuildServerURL, "the BuildServerURL should be equals")
				assert.Equal(t, "jenkins", *got.BuildServerProvider, "the BuildServerProvider should be equals")
				assert.Equal(t, "jenkins", *got.ContinuousIntegrationProvider, "the ContinuousIntegrationProvider should be equals")
				assert.Equal(t, "gitflow", *got.WorkflowType, "the WorkflowType should be equals")
				assert.Equal(t, 80.0, *got.CodeCoveragePullRequestThreshold, "the CodeCoveragePullRequestThreshold should be equals")
			}
		})
	}
}

func TestConfiguration_Get(t *testing.T) {
	type args struct {
		id string
	}

	type expects struct {
		errorLog string
		infoLog  string
		error    error
		config   models.Configuration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test Get - Config obtained successfully, should return it",
			args: args{
				id: "fury_repo-name",
			},
			wantErr: false,
		},
		{
			name: "test Get - Error checking configuration existence, should return an Error",
			args: args{
				id: "fury_repo-name",
			},
			expects: expects{
				error: errors.New("record not found"),
			},
			wantErr: true,
		},
		{
			name: "test Get - gorm Error record Not Found",
			args: args{
				id: "fury_repo-name",
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctl)

			logger := interfaces.NewMockLogger(ctl)

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.expects.error).
				AnyTimes()

			s := &Configuration{
				SQL:    sqlStorage,
				Logger: logger,
			}

			_, err := s.Get(tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestConfiguration_Update(t *testing.T) {
	type workflowFuncResult struct {
		setWorkflow   error
		unsetWorkflow error
	}

	type builderFuncResult struct {
		createJob error
		deleteJob error
	}

	type clientsResult struct {
		furyClient     error
		workflowClient workflowFuncResult
		builderClient  builderFuncResult
		netRPClient    error
		melicovClient  error
		sqlClient      error
	}

	type workflowClientFuncExecution struct {
		execute       bool
		setWorkflow   bool
		unsetWorkflow bool
	}

	type builderClientFuncExecution struct {
		execute   bool
		createJob bool
		deleteJob bool
	}

	type clientsExecution struct {
		furyClient     bool
		workflowClient workflowClientFuncExecution
		builderClient  builderClientFuncExecution
		netRPClient    bool
		melicovClient  bool
	}

	type args struct {
		putRequestPayload *models.PutRequestPayload
		config            models.Configuration
		execute           clientsExecution
	}

	type expects struct {
		errorLog      string
		infoLog       string
		error         error
		config        models.Configuration
		clientsResult clientsResult
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	payload := models.PutRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	tests := []struct {
		name          string
		args          args
		wantErr       bool
		expects       expects
		clientsResult clientsResult
	}{
		{
			name: "test Update - updated and saved successfully",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test Update - pull request coverage threshold update fails. We do nothing",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					melicovClient: errors.New("some error"),
				},
			},
			wantErr: false,
		},
		{
			name: "test Update - save configuration fails, should return an error",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					melicovClient: errors.New("some error"),
				},
			},
			wantErr: false,
		},
		{
			name: "test Update - repository config not found, should return and error",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					sqlClient: errors.New("not found"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Update - setWorkflow fails, should return and error",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					workflowClient: workflowFuncResult{
						setWorkflow: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Update - ci's job update failed, we must rollback the workflow and return an error",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					builderClient: builderFuncResult{
						createJob: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Update - ci's job update failed, we must rollback the workflow. Roll back fails",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					builderClient: builderFuncResult{
						createJob: errors.New("some error"),
					},
					workflowClient: workflowFuncResult{
						unsetWorkflow: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Update - application technology update failed, we must rollback the workflow and CI pipeline and return an error",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					furyClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Update - application technology update failed, we must rollback the workflow and CI pipeline. Workflow's rollback fails",
			args: args{
				putRequestPayload: &payload,
				config:            repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					furyClient: errors.New("some error"),
					workflowClient: workflowFuncResult{
						setWorkflow: errors.New("some error"),},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Release Process Clients
			furyClient := interfaces.NewMockFuryClient(ctl)
			builderClient := interfaces.NewMockBuilderClient(ctl)
			workflowClient := interfaces.NewMockWorkflowClient(ctl)
			melicovClient := interfaces.NewMockMelicovClient(ctl)
			netRPClient := interfaces.NewMockNetRPClient(ctl)

			//SQL Clients
			sqlStorage := interfaces.NewMockSQLStorage(ctl)

			//Logger
			logger := interfaces.NewMockLogger(ctl)

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
					return &tt.expects.config
				}).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			//SQLStorageClient

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			sqlStorage.EXPECT().
				Update(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			sqlStorage.EXPECT().
				DeleteFromRequireStatusChecksByConfigurationID(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			//FuryClient

			//GetApplicationData
			furyClient.EXPECT().
				GetApplicationData(gomock.Any()).
				Return(tt.expects.clientsResult.furyClient).
				AnyTimes()

			//UpdateApplicationTechnology
			furyClient.EXPECT().
				UpdateApplicationTechnology(gomock.Any()).
				Return(tt.expects.clientsResult.furyClient).
				AnyTimes()

			//WorkflowClient

			//SetWorkflow
			workflowClient.EXPECT().
				SetWorkflow(gomock.Any()).
				Return(tt.expects.clientsResult.workflowClient.setWorkflow).
				AnyTimes()

			//UnsetWorkflow
			workflowClient.EXPECT().
				UnSetWorkflow(gomock.Any()).
				Return(tt.expects.clientsResult.workflowClient.unsetWorkflow).
				AnyTimes()

			//Builder Client

			//Create Job
			builderClient.EXPECT().
				CreateJob(gomock.Any()).
				Return(tt.expects.clientsResult.builderClient.createJob).
				AnyTimes()

			//Delete Job
			builderClient.EXPECT().
				DeleteJob(gomock.Any()).
				Return(tt.expects.clientsResult.builderClient.deleteJob).
				AnyTimes()

			//Melicov Client

			melicovClient.EXPECT().
				SetPullRequestThreshold(gomock.Any()).
				Return(tt.expects.clientsResult.melicovClient).
				AnyTimes()

			melicovClient.EXPECT().
				UpdatePullRequestThreshold(gomock.Any()).
				Return(tt.expects.clientsResult.melicovClient).
				AnyTimes()

			//NetRP Client

			netRPClient.EXPECT().
				EnableReleaseProcess(gomock.Any()).
				Return(tt.expects.clientsResult.netRPClient).
				AnyTimes()

			s := &Configuration{
				SQL:            sqlStorage,
				Logger:         logger,
				FuryClient:     furyClient,
				WorkflowClient: workflowClient,
				BuilderClient:  builderClient,
				MelicovClient:  melicovClient,
				NetRPClient:    netRPClient,
			}

			_, err := s.Update(tt.args.putRequestPayload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpdate_SaveNewConfigurationFails(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PutRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.Repository.RequireStatusChecks = statusList

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)

	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(nil).
			AnyTimes(),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		sqlStorage.EXPECT().
			DeleteFromRequireStatusChecksByConfigurationID(gomock.Any()).
			Return(nil).
			AnyTimes(),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		//UpdateApplicationTechnology
		furyClient.EXPECT().
			UpdateApplicationTechnology(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			UpdatePullRequestThreshold(gomock.Any()).
			Return(nil).
			AnyTimes(),

		sqlStorage.EXPECT().
			Update(gomock.Any()).
			Return(errors.New("some db error")).
			AnyTimes(),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
	}

	_, err := s.Update(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Update() error = %v", err)
		return
	}

}

func TestTechnologyConfigurationUpdate_SetWorkflowFailsWhenRollbackWasDone(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PutRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(nil).
			AnyTimes(),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		sqlStorage.EXPECT().
			DeleteFromRequireStatusChecksByConfigurationID(gomock.Any()).
			Return(nil).
			AnyTimes(),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		//UpdateApplicationTechnology
		furyClient.EXPECT().
			UpdateApplicationTechnology(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			AnyTimes(),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
	}

	_, err := s.Update(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Update() error = %v", err)
		return
	}

}

func TestTechnologyConfigurationUpdate_DeleteReqStatusChecksFailsWhenRollbackWasDone(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PutRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(nil).
			AnyTimes(),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		sqlStorage.EXPECT().
			DeleteFromRequireStatusChecksByConfigurationID(gomock.Any()).
			Return(errors.New("some error")).
			AnyTimes(),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
	}

	_, err := s.Update(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Update() error = %v", err)
		return
	}

}

func TestTechnologyConfigurationUpdate_CreateJobFailsWhenRollbackWasDone(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PutRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(nil).
			AnyTimes(),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		sqlStorage.EXPECT().
			DeleteFromRequireStatusChecksByConfigurationID(gomock.Any()).
			Return(nil).
			AnyTimes(),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		//UpdateApplicationTechnology
		furyClient.EXPECT().
			UpdateApplicationTechnology(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(errors.New("some error")).
			AnyTimes(),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
	}

	_, err := s.Update(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Update() error = %v", err)
		return
	}

}

func TestConfiguration_Delete(t *testing.T) {
	type workflowFuncResult struct {
		setWorkflow   error
		unsetWorkflow error
	}

	type builderFuncResult struct {
		createJob error
		deleteJob error
	}

	type sqlClientFuncResult struct {
		getBy  error
		delete error
	}

	type clientsResult struct {
		furyClient     error
		workflowClient workflowFuncResult
		builderClient  builderFuncResult
		netRPClient    error
		melicovClient  error
		sqlClient      sqlClientFuncResult
	}

	type workflowClientFuncExecution struct {
		setWorkflow   bool
		unsetWorkflow bool
	}

	type builderClientFuncExecution struct {
		createJob bool
		deleteJob bool
	}

	type clientsExecution struct {
		furyClient     bool
		workflowClient workflowClientFuncExecution
		builderClient  builderClientFuncExecution
		netRPClient    bool
		melicovClient  bool
	}

	type args struct {
		repository string
		config     models.Configuration
		execute    clientsExecution
	}

	type expects struct {
		errorLog      string
		infoLog       string
		error         error
		config        models.Configuration
		clientsResult clientsResult
	}

	repoConfigOK := models.Configuration{
		ID: utils.Stringify("fury_repo-name"),
	}

	tests := []struct {
		name          string
		args          args
		wantErr       bool
		expects       expects
		clientsResult clientsResult
	}{
		{
			name: "test Delete - configuration deleted successfully",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test Delete - repository configuration not found, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					sqlClient: sqlClientFuncResult{
						getBy: errors.New("configuration not found"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Delete - unsetWorkflow fails, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					workflowClient: workflowFuncResult{
						unsetWorkflow: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Delete - delete CI Jobs fails, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					builderClient: builderFuncResult{
						deleteJob: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test Delete - disable netRP configuration fails, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					netRPClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Delete - disable Fury release process field fails, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					furyClient: errors.New("some error"),
				},
			},
			wantErr: true,
		},
		{
			name: "test Delete - delete configuration from db fails, should return an error",
			args: args{
				repository: "fury_repo-name",
				config:     repoConfigOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					sqlClient: sqlClientFuncResult{
						delete: errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Release Process Clients
			furyClient := interfaces.NewMockFuryClient(ctl)
			builderClient := interfaces.NewMockBuilderClient(ctl)
			workflowClient := interfaces.NewMockWorkflowClient(ctl)
			melicovClient := interfaces.NewMockMelicovClient(ctl)
			netRPClient := interfaces.NewMockNetRPClient(ctl)

			//SQL Clients
			sqlStorage := interfaces.NewMockSQLStorage(ctl)

			//Logger
			logger := interfaces.NewMockLogger(ctl)

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
					return &tt.expects.config
				}).
				Return(tt.expects.clientsResult.sqlClient.getBy).
				AnyTimes()

			//SQLStorageClient

			sqlStorage.EXPECT().
				Delete(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient.delete).
				AnyTimes()

			//WorkflowClient

			//UnsetWorkflow
			workflowClient.EXPECT().
				UnSetWorkflow(gomock.Any()).
				Return(tt.expects.clientsResult.workflowClient.unsetWorkflow).
				AnyTimes()

			//Builder Client

			//Delete Job
			builderClient.EXPECT().
				DeleteJob(gomock.Any()).
				Return(tt.expects.clientsResult.builderClient.deleteJob).
				AnyTimes()

			//NetRP Client

			netRPClient.EXPECT().
				DisableReleaseProcess(gomock.Any()).
				Return(tt.expects.clientsResult.netRPClient).
				AnyTimes()

			furyClient.EXPECT().
				DisableReleaseProcessField(gomock.Any()).
				Return(tt.expects.clientsResult.furyClient).
				AnyTimes()

			s := &Configuration{
				SQL:            sqlStorage,
				Logger:         logger,
				FuryClient:     furyClient,
				WorkflowClient: workflowClient,
				BuilderClient:  builderClient,
				MelicovClient:  melicovClient,
				NetRPClient:    netRPClient,
			}

			err := s.Delete(tt.args.repository)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCreateConfiguration_EnableRPOnFuryFails_RollBackWasDoneOK(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			SetPullRequestThreshold(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			EnableReleaseProcess(gomock.Any()).
			Return(nil).
			Times(1),

		furyClient.EXPECT().
			EnableReleaseProcessField(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//RollBack

		//SetWorkflow
		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			DeleteJob(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			DisableReleaseProcess(gomock.Any()).
			Return(nil).
			Times(1),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		NetRPClient:    netRPClient,
		MelicovClient:  melicovClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}

func TestCreateConfiguration_EnableRPOnFuryFails_RollFailsWhenUnsetWorkflowWasDoneK(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			SetPullRequestThreshold(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			EnableReleaseProcess(gomock.Any()).
			Return(nil).
			Times(1),

		furyClient.EXPECT().
			EnableReleaseProcessField(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//RollBack

		//SetWorkflow
		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		NetRPClient:    netRPClient,
		MelicovClient:  melicovClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}

func TestCreateConfiguration_EnableRPOnFuryFails_BuilderRollBackFails(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			SetPullRequestThreshold(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			EnableReleaseProcess(gomock.Any()).
			Return(nil).
			Times(1),

		furyClient.EXPECT().
			EnableReleaseProcessField(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//RollBack

		//SetWorkflow
		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			DeleteJob(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		NetRPClient:    netRPClient,
		MelicovClient:  melicovClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}

func TestCreateConfiguration_EnableRPOnFuryFails_NetRPRollBackFails(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			SetPullRequestThreshold(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			EnableReleaseProcess(gomock.Any()).
			Return(nil).
			Times(1),

		furyClient.EXPECT().
			EnableReleaseProcessField(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//RollBack

		//SetWorkflow
		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			DeleteJob(gomock.Any()).
			Return(nil).
			Times(1),

		netRPClient.EXPECT().
			DisableReleaseProcess(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		NetRPClient:    netRPClient,
		MelicovClient:  melicovClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}

func TestCreateConfiguration_CreateCIJobsFails_RollbackWasDoneOK(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		//RollBack

		//SetWorkflow
		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		NetRPClient:    netRPClient,
		MelicovClient:  melicovClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}

func TestCreateConfiguration_SetPullRequestThresholdFails_NothingWasDone(t *testing.T) {
	wantErr := true
	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}
	codeCoverageThreadhold := 80.0

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	payload := models.PostRequestPayload{}

	payload.Repository.Name = utils.Stringify("fury_repo-name")
	payload.Repository.RequireStatusChecks = statusList
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	repoConfigOK := models.Configuration{
		ID:                               utils.Stringify("fury_repo-name"),
		ApplicationName:                  utils.Stringify("repo-name"),
		Technology:                       utils.Stringify("java"),
		RepositoryURL:                    utils.Stringify("http://github.com/fury_repo-name"),
		ContinuousIntegrationURL:         utils.Stringify("https://rp-ci.furycloud.io/job/repo-name/"),
		BuildServerURL:                   utils.Stringify("https://rp-builds.furycloud.io/job/repo-name/"),
		BuildServerProvider:              utils.Stringify("jenkins"),
		ContinuousIntegrationProvider:    utils.Stringify("jenkins"),
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		RepositoryStatusChecks:           reqChecks,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	//Release Process Clients
	furyClient := interfaces.NewMockFuryClient(ctl)
	builderClient := interfaces.NewMockBuilderClient(ctl)
	workflowClient := interfaces.NewMockWorkflowClient(ctl)
	melicovClient := interfaces.NewMockMelicovClient(ctl)
	netRPClient := interfaces.NewMockNetRPClient(ctl)

	//SQL Clients
	sqlStorage := interfaces.NewMockSQLStorage(ctl)

	//Logger
	logger := interfaces.NewMockLogger(ctl)

	gomock.InOrder(
		sqlStorage.EXPECT().
			GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
				return &repoConfigOK
			}).
			Return(gorm.ErrRecordNotFound).
			Times(1),

		furyClient.EXPECT().
			GetApplicationData(gomock.Any()).
			Return(nil).
			Times(1),

		//SetWorkflow
		workflowClient.EXPECT().
			SetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			CreateJob(gomock.Any()).
			Return(nil).
			Times(1),

		melicovClient.EXPECT().
			SetPullRequestThreshold(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		netRPClient.EXPECT().
			EnableReleaseProcess(gomock.Any()).
			Return(errors.New("some error")).
			Times(1),

		workflowClient.EXPECT().
			UnSetWorkflow(gomock.Any()).
			Return(nil).
			Times(1),

		builderClient.EXPECT().
			DeleteJob(gomock.Any()).
			Return(nil).
			Times(1),
	)

	s := &Configuration{
		SQL:            sqlStorage,
		Logger:         logger,
		FuryClient:     furyClient,
		WorkflowClient: workflowClient,
		BuilderClient:  builderClient,
		MelicovClient:  melicovClient,
		NetRPClient:    netRPClient,
	}

	_, err := s.Create(&payload)

	if (err != nil) != wantErr {
		t.Errorf("Configuration.Create() error = %v", err)
		return
	}

}
