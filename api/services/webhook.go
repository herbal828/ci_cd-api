package services

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/herbal828/ci_cd-api/api/clients"
	"github.com/herbal828/ci_cd-api/api/models"
	"github.com/herbal828/ci_cd-api/api/models/webhook"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"github.com/herbal828/ci_cd-api/api/utils"
	"github.com/herbal828/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"io/ioutil"
)

type WebhookService interface {
	CreateWebhook(ctx *gin.Context, webhookEvent string) (*webhook.Webhook, apierrors.ApiError)
	ProcessStatusWebhook(ctx utils.HTTPContext, conf *models.Configuration) (*webhook.Webhook, apierrors.ApiError)
}

//Webhook represents the WebhookService layer
//It has an instance of a DBClient layer and
//A github client instance
type Webhook struct {
	SQL          storage.SQLStorage
	GithubClient clients.GithubClient
}

//NewConfigurationSeNewWebhookServicervice initializes a WebhookService
func NewWebhookService(sql storage.SQLStorage) *Webhook {
	return &Webhook{
		SQL:          sql,
		GithubClient: clients.NewGithubClient(),
	}
}

//CreateWebhook creates a new webhook for the given repository
//It could returns
//	200OK in case of a success processing the creation
//	400BadRequest in case of an error parsing the request payload
//	500InternalServerError in case of an internal error procesing the creation
func (s *Webhook) CreateWebhook(ctx *gin.Context, webhookEvent string) (*webhook.Webhook, apierrors.ApiError) {

	// Read the content
	var bodyBytes []byte
	if ctx.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(ctx.Request.Body)
	}
	// Restore the io.ReadCloser to its original state
	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	//validate that the repository comes in the payload
	var ghPayload webhook.GithubWebhookStandardPayload
	if err := json.Unmarshal(bodyBytes, &ghPayload); err != nil {
		return nil, apierrors.NewBadRequestApiError("invalid github webhook payload")
	}

	repository := ghPayload.Repository.Name

	var config models.Configuration

	//Validates that the repository has a ci cd configuration
	if err := s.SQL.GetBy(&config, "id = ?", *repository); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("repository dosn't have a ci-cd configuration")
		} else {
			return nil, apierrors.NewInternalServerApiError("error checking configuration existence", err)
		}
	}

	var wh webhook.Webhook

	if webhookEvent == "" {
		return nil, apierrors.NewBadRequestApiError("x-github-event is null")
	}

	switch webhookEvent {
	case "status":
		wh, err := s.ProcessStatusWebhook(ctx, &config)
		if err != nil {
			return nil, err
		}
		return wh, nil
	case "pull_request_review":
	case "issue_comment":
	case "pull_request":
		wh, err := s.ProcessStatusWebhook(ctx, &config)
		if err != nil {
			return nil, err
		}
		return wh, nil
	case "create":
	default:
		return nil, apierrors.NewBadRequestApiError("Event not supported yet")
	}
	return &wh, nil
}

//ProcessStatusWebhook process
func (s *Webhook) ProcessStatusWebhook(ctx utils.HTTPContext, conf *models.Configuration) (*webhook.Webhook, apierrors.ApiError) {

	var statusWH webhook.Status
	var wh webhook.Webhook

	if err := ctx.BindJSON(&statusWH); err != nil {
		return nil, apierrors.NewBadRequestApiError("invalid status webhook payload")
	}

	contextAllowed := utils.Contains(conf.RepositoryStatusChecks, statusWH.Context)
	if !contextAllowed {
		return nil, apierrors.NewBadRequestApiError("Context not configured for the repository")
	}

	//Build a ID to identify a unique webhook
	shBaseID := statusWH.Repository.FullName + statusWH.Sha + statusWH.Context + statusWH.State
	statusWebhookID := utils.Stringify(utils.GetMD5Hash(shBaseID))

	//Search the status webhook into database
	if err := s.SQL.GetBy(&wh, "id = ?", &statusWebhookID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking configuration existence")
		}

		//Fill every field in the webhook
		wh.ID = statusWebhookID
		wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(ctx.GetHeader("X-GitHub-Event"))
		wh.GithubRepositoryName = utils.Stringify(statusWH.Repository.FullName)
		wh.SenderName = utils.Stringify(statusWH.Sender.Login)
		wh.WebhookCreateAt = statusWH.CreatedAt
		wh.WebhookUpdated = statusWH.UpdatedAt
		wh.State = utils.Stringify(statusWH.State)
		wh.Context = utils.Stringify(statusWH.Context)
		wh.Sha = utils.Stringify(statusWH.Sha)
		wh.Description = utils.Stringify(statusWH.Description)

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new status webhook", err)
		}

	} else { //If webhook already exists then return it
		return nil, apierrors.NewConflictApiError("Resource Already exists")
	}

	return &wh, nil
}