package controllers

import (
	"github.com/herbal828/ci_cd-api/api/services"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"github.com/herbal828/ci_cd-api/api/utils"
	"github.com/herbal828/ci_cd-api/api/utils/apierrors"
	"net/http"
)

type Webhook struct {
	Service   services.WebhookService
}

//NewWebhookController initializes a WebhookController
func NewWebhookController(sql storage.SQLStorage) *Webhook {
	return &Webhook{
		Service: services.NewWebhookService(sql),
	}
}

//Create creates a new configuration for the given repository
//It could returns
//	200OK in case of a success processing the creation
//	400BadRequest in case of an error parsing the request payload
//	500InternalServerError in case of an internal error procesing the creation
func (c *Webhook) CreateWebhook(ctx utils.HTTPContext) {

	whook, createWHErr := c.Service.CreateWebhook(ctx)
	if createWHErr != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			apierrors.NewInternalServerApiError("something was wrong creating a new webhook", createWHErr),
		)
		return
	}

	ctx.JSON(http.StatusOK, whook.Marshall())
}


