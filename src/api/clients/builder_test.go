package clients

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilderClient_CreateJob(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config *models.Configuration
		body   map[string]interface{}
	}

	type expects struct {
		errorLog string
		infoLog  string
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "test nil ApplicationName should return an error",
			args: args{
				config: &models.Configuration{
					ID:              nil,
					Technology:      nil,
					ApplicationName: nil,
					RepositoryURL:   nil,
				},
			},
			expects: expects{
				errorLog: "one of the create ci and build-server parameters are null",
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "test nil Technology should return an error",
			args: args{
				config: &models.Configuration{
					Technology: nil,
				},
			},
			expects: expects{
				errorLog: "one of the create ci and build-server parameters are null",
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "test nil RepositoryURL should return an error",
			args: args{
				config: &models.Configuration{
					RepositoryURL: nil,
				},
			},
			expects: expects{
				errorLog: "one of the create ci and build-server parameters are null",
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "test Post return an error",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-name"),
					RepositoryURL:   utils.Stringify("http://github.com/repos/fury_repository-name"),
					ApplicationName: utils.Stringify("repository-name"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repository-name"),
					"repository_url": *utils.Stringify("http://github.com/repos/fury_repository-name"),
					"technology":     *utils.Stringify("java"),
				},
			},
			restResponse: restResponse{
				mockError: errors.New("Some Bad Error"),
			},
			expects: expects{
				errorLog: "creating ci and build-server",
			},
			wantErr: true,
		},
		{
			name: "test Post OK - ci and build-server jobs created successfully",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-01"),
					RepositoryURL:   utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
					ApplicationName: utils.Stringify("repo-01"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repo-01"),
					"technology":     *utils.Stringify("java"),
					"repository_url": *utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
				},
			},
			restResponse: restResponse{
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"message":  "mensaje",
					"url":      "java",
					"provider": "java",
				}),
			},
			expects: expects{
				infoLog: "ci and build-server job created successfully",
			},
			wantErr: false,
		},
		{
			name: "test Post Fails with an Marshalling error",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-01"),
					RepositoryURL:   utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
					ApplicationName: utils.Stringify("repo-01"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repo-01"),
					"technology":     *utils.Stringify("java"),
					"repository_url": *utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
				},
			},
			restResponse: restResponse{
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"message":  1,
					"url":      "java",
					"provider": "java",
				}),
			},
			expects: expects{
				errorLog: "marshalling ci-proxy create job response",
			},
			wantErr: true,
		},
		{
			name: "test Post return 400",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-name"),
					RepositoryURL:   utils.Stringify("http://github.com/repos/fury_repository-name"),
					ApplicationName: utils.Stringify("repository-name"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repository-name"),
					"technology":     *utils.Stringify("java"),
					"repository_url": *utils.Stringify("http://github.com/repos/fury_repository-name"),
				},
			},
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			expects: expects{
				infoLog: "something was wrong creating ci and build-server",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)
			logger := interfaces.NewMockLogger(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), tt.args.body).
				Return(response).
				AnyTimes()

			logger.EXPECT().
				CreateTag(gomock.Any(), gomock.Any()).
				DoAndReturn(func(k string, v interface{}) string {
					return fmt.Sprintf(k, v)
				}).
				AnyTimes()

			logger.EXPECT().
				Error(tt.expects.errorLog, gomock.Any(), gomock.Any()).
				AnyTimes()

			logger.EXPECT().
				Info(tt.expects.infoLog, gomock.Any()).
				AnyTimes()

			c := &builderClient{
				Client: client,
				Logger: logger,
			}

			if err := c.CreateJob(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("BuilderClient.CreateJob() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Equal(t, "fury_repo-01", *tt.args.config.ID, "the IDs should be equals")
				assert.Equal(t, "java", *tt.args.config.Technology, "the techs should be equals")
				assert.Equal(t, "repo-01", *tt.args.config.ApplicationName, "the apps should be equals")
				assert.Equal(t, "https://github.com/mercadolibre/fury_repo-01", *tt.args.config.RepositoryURL, "the repo URLs should be equals")
			}
		})
	}
}

func TestBuilderClient_DeleteJob(t *testing.T) {

	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config *models.Configuration
		body   map[string]interface{}
	}

	type expects struct {
		errorLog string
		infoLog  string
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "test nil ApplicationName should return an error",
			args: args{
				config: &models.Configuration{
					ID:              nil,
					Technology:      nil,
					ApplicationName: nil,
					RepositoryURL:   nil,
				},
			},
			expects: expects{
				errorLog: "application name parameter is null",
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "test Delete return an error",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-name"),
					RepositoryURL:   utils.Stringify("http://github.com/repos/fury_repository-name"),
					ApplicationName: utils.Stringify("repository-name"),
					Technology:      utils.Stringify("java"),
				},
			},
			restResponse: restResponse{
				mockError: errors.New("Some Bad Error"),
			},
			expects: expects{
				errorLog: "deleting ci and build-server",
			},
			wantErr: true,
		},
		{
			name: "test Delete OK - ci and build-server jobs deleted successfully",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-01"),
					RepositoryURL:   utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
					ApplicationName: utils.Stringify("repo-01"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repo-01"),
					"technology":     *utils.Stringify("java"),
					"repository_url": *utils.Stringify("https://github.com/mercadolibre/fury_repo-01"),
				},
			},
			restResponse: restResponse{
				mockStatusCode: 200,
			},
			expects: expects{
				infoLog: "ci and build-server job deleted successfully",
			},
			wantErr: false,
		},
		{
			name: "test Delete return 400",
			args: args{
				config: &models.Configuration{
					ID:              utils.Stringify("fury_repo-name"),
					RepositoryURL:   utils.Stringify("http://github.com/repos/fury_repository-name"),
					ApplicationName: utils.Stringify("repository-name"),
					Technology:      utils.Stringify("java"),
				},
				body: map[string]interface{}{
					"name":           *utils.Stringify("repository-name"),
					"technology":     *utils.Stringify("java"),
					"repository_url": *utils.Stringify("http://github.com/repos/fury_repository-name"),
				},
			},
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			expects: expects{
				infoLog: "something was wrong deleting ci and build-server",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)
			logger := interfaces.NewMockLogger(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Delete(gomock.Any()).
				Return(response).
				AnyTimes()

			logger.EXPECT().
				CreateTag(gomock.Any(), gomock.Any()).
				DoAndReturn(func(k string, v interface{}) string {
					return fmt.Sprintf(k, v)
				}).
				AnyTimes()

			logger.EXPECT().
				Error(tt.expects.errorLog, gomock.Any(), gomock.Any()).
				AnyTimes()

			logger.EXPECT().
				Info(tt.expects.infoLog, gomock.Any()).
				AnyTimes()

			c := &builderClient{
				Client: client,
				Logger: logger,
			}

			if err := c.DeleteJob(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("BuilderClient.DeleteJob() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Equal(t, "fury_repo-01", *tt.args.config.ID, "the IDs should be equals")
				assert.Equal(t, "java", *tt.args.config.Technology, "the techs should be equals")
				assert.Equal(t, "repo-01", *tt.args.config.ApplicationName, "the apps should be equals")
				assert.Equal(t, "https://github.com/mercadolibre/fury_repo-01", *tt.args.config.RepositoryURL, "the repo URLs should be equals")
			}
		})
	}
}
