package controllers

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/jinzhu/gorm"
	"net/http"
	"testing"
)

func TestConfiguration_Create(t *testing.T) {

	type resp struct {
		httpStatusCode int
	}

	type args struct {
		parseMock func(i interface{}) error
	}

	type expects struct {
		error    error
		config   *models.Configuration
		ctxError error
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
		resp    resp
		expects expects
		wantErr bool
	}{
		{
			name: "test Create, create repository release process configuration successfully",
			args: args{
				parseMock: func(i interface{}) error {
					fmt.Println("#")
					return nil
				},
			},
			resp: resp{
				httpStatusCode: 200,
			},
			expects: expects{
				config:   &repoConfigOK,
				ctxError: nil,
			},
			wantErr: false,
		},
		{
			name: "test Create, BindJson Fails",
			args: args{
				parseMock: func(i interface{}) error {
					fmt.Println("#")
					return errors.New("invalid configuration request payload")
				},
			},
			resp: resp{
				httpStatusCode: 400,
			},
			expects: expects{
				config:   &repoConfigOK,
				ctxError: apierrors.NewBadRequestApiError("invalid configuration request payload"),
			},
			wantErr: true,
		},
		{
			name: "test Create, create configurations Fails, should return an error",
			args: args{
				parseMock: func(i interface{}) error {
					fmt.Println("#")
					return nil
				},
			},
			resp: resp{
				httpStatusCode: 500,
			},
			expects: expects{
				config:   &repoConfigOK,
				error:    errors.New("some error"),
				ctxError: apierrors.NewInternalServerApiError("something was wrong creating a new configuration", errors.New("some error")),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Configuration Service
			service := interfaces.NewMockConfigurationService(ctl)
			logger := interfaces.NewMockLogger(ctl)
			ctx := interfaces.NewMockHTTPContext(ctl)

			service.
				EXPECT().
				Create(gomock.Any()).
				Return(tt.expects.config, tt.expects.error).
				AnyTimes()

			c := &Configuration{
				Service: service,
				Logger:  logger,
			}

			ctx.
				EXPECT().
				BindJSON(gomock.Any()).
				DoAndReturn(tt.args.parseMock).
				AnyTimes()

			if tt.wantErr {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.ctxError)).
					AnyTimes()
			} else {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.config.Marshall())).
					AnyTimes()
			}

			c.Create(ctx)

		})
	}
}

func TestConfiguration_Get(t *testing.T) {

	type resp struct {
		httpStatusCode int
	}

	type args struct {
		parseMock func(i interface{}) error
		param     string
	}

	type expects struct {
		error    error
		config   *models.Configuration
		ctxError error
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
		resp    resp
		expects expects
		wantErr bool
	}{
		{
			name: "test Get, repository release process configuration obtained successfully",
			args: args{

				param: "fury_repo-name",
			},
			resp: resp{
				httpStatusCode: 200,
			},
			expects: expects{
				config:   &repoConfigOK,
				ctxError: nil,
			},
			wantErr: false,
		},
		{
			name: "test Get, error when obtaining the configuration, should return an error",
			resp: resp{
				httpStatusCode: 500,
			},
			expects: expects{
				error:    errors.New("some db error"),
				config:   &repoConfigOK,
				ctxError: apierrors.NewInternalServerApiError("something was wrong getting the configuration for ", errors.New("some db error")),
			},
			wantErr: true,
		},
		{
			name: "test Get, configuration not found ",
			resp: resp{
				httpStatusCode: http.StatusNotFound,
			},
			expects: expects{
				error:    gorm.ErrRecordNotFound,
				config:   &repoConfigOK,
				ctxError: apierrors.NewNotFoundApiError("configuration for repository  not found"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Configuration Service
			service := interfaces.NewMockConfigurationService(ctl)
			logger := interfaces.NewMockLogger(ctl)
			ctx := interfaces.NewMockHTTPContext(ctl)

			service.
				EXPECT().
				Get(gomock.Any()).
				Return(tt.expects.config, tt.expects.error).
				AnyTimes()

			c := &Configuration{
				Service: service,
				Logger:  logger,
			}

			ctx.
				EXPECT().
				Param(gomock.Any()).
				Return(tt.args.param).
				AnyTimes()

			if tt.wantErr {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.ctxError)).
					AnyTimes()
			} else {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.config.Marshall())).
					AnyTimes()
			}

			c.Show(ctx)

		})
	}
}

func TestConfiguration_Update(t *testing.T) {

	type resp struct {
		httpStatusCode int
	}

	type args struct {
		parseMock func(i interface{}) error
		param     string
	}

	type expects struct {
		error    error
		config   *models.Configuration
		ctxError error
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage"}
	codeCoverageThreadhold := 70.0

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
	payload.Repository.RequireStatusChecks = statusList
	payload.Fury.Technology = utils.Stringify("java-gradle")
	payload.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold

	tests := []struct {
		name    string
		args    args
		resp    resp
		expects expects
		wantErr bool
	}{
		{
			name: "test Update, configuration updated successfully",
			args: args{
				parseMock: func(i interface{}) error {
					fmt.Println("#")
					return nil
				},
				param: "fury_repo-name",
			},

			resp: resp{
				httpStatusCode: 200,
			},
			expects: expects{
				config:   &repoConfigOK,
				ctxError: nil,
			},
			wantErr: false,
		},
		{
			name: "test Update, BindJson fails, should return an error",
			args: args{
				parseMock: func(i interface{}) error {
					fmt.Println("#")
					return errors.New("invalid configuration request payload")
				},
			},
			resp: resp{
				httpStatusCode: 400,
			},
			expects: expects{
				config:   &repoConfigOK,
				ctxError: apierrors.NewBadRequestApiError("invalid configuration request payload"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Configuration Service
			service := interfaces.NewMockConfigurationService(ctl)
			logger := interfaces.NewMockLogger(ctl)
			ctx := interfaces.NewMockHTTPContext(ctl)

			service.
				EXPECT().
				Update(gomock.Any()).
				Return(tt.expects.config, tt.expects.error).
				AnyTimes()

			c := &Configuration{
				Service: service,
				Logger:  logger,
			}

			ctx.
				EXPECT().
				BindJSON(gomock.Any()).
				DoAndReturn(tt.args.parseMock).
				AnyTimes()

			ctx.
				EXPECT().
				Param(gomock.Any()).
				Return(tt.args.param).
				AnyTimes()

			if tt.wantErr {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.ctxError)).
					AnyTimes()
			} else {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.config.Marshall())).
					AnyTimes()
			}

			c.Update(ctx)

		})
	}
}

func TestConfiguration_Delete(t *testing.T) {

	type resp struct {
		httpStatusCode int
	}

	type args struct {
		parseMock func(i interface{}) error
		param     string
	}

	type expects struct {
		error    error
		ctxError error
	}

	tests := []struct {
		name    string
		args    args
		resp    resp
		expects expects
		wantErr bool
	}{
		{
			name: "test Delete, configuration deleted successfully",
			args: args{
				param: "fury_repo-name",
			},

			resp: resp{
				httpStatusCode: 204,
			},
			expects: expects{
				ctxError: nil,
			},
			wantErr: false,
		},
		{
			name: "test Delete, configuration not found, should return an error",
			args: args{
				param: "fury_repo-name",
			},

			resp: resp{
				httpStatusCode: 404,
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
				ctxError: apierrors.NewNotFoundApiError("configuration for repository fury_repo-name not found"),
			},
			wantErr: true,
		},
		{
			name: "test Delete, configuration deletes fail, should return an error",
			args: args{
				param: "fury_repo-name",
			},

			resp: resp{
				httpStatusCode: 500,
			},
			expects: expects{
				error: errors.New("some error"),
				ctxError: apierrors.NewInternalServerApiError("something was wrong getting the configuration for fury_repo-name", errors.New("some error")),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			//Configuration Service
			service := interfaces.NewMockConfigurationService(ctl)
			logger := interfaces.NewMockLogger(ctl)
			ctx := interfaces.NewMockHTTPContext(ctl)

			service.
				EXPECT().
				Delete(gomock.Any()).
				Return(tt.expects.error).
				AnyTimes()

			c := &Configuration{
				Service: service,
				Logger:  logger,
			}

			ctx.
				EXPECT().
				BindJSON(gomock.Any()).
				DoAndReturn(tt.args.parseMock).
				AnyTimes()

			ctx.
				EXPECT().
				Param(gomock.Any()).
				Return(tt.args.param).
				AnyTimes()

			if tt.wantErr {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.ctxError)).
					AnyTimes()
			} else {
				ctx.
					EXPECT().
					JSON(gomock.Eq(tt.resp.httpStatusCode), gomock.Eq(tt.expects.error)).
					AnyTimes()
			}

			c.Delete(ctx)

		})
	}
}
