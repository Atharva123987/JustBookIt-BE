package controller

import (
	"errors"
	"fmt"
	"net/http"
	apiresponse "ticketBookingSystem/internal/apiResponse"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type WebhookController interface {
	StripeWebhook(ctx *gin.Context)
}

var StripeWebhookCheckoutSessionCompleted string = "checkout.session.completed"
var StripeWebhookInvoicePaid = "invoice.paid"

type webhookController struct{}

func (wc *webhookController) StripeWebhook(ctx *gin.Context) {
	var data map[string]interface{}
	err := ctx.BindJSON(&data)
	if err != nil {
		apiresponse.NewErrorResponse(ctx, err, http.StatusBadRequest)
		return
	}
	webhookType, ok := data["type"].(string)
	if !ok || webhookType == "" {
		apiresponse.NewErrorResponse(ctx, errors.New("type field not found or invalid format"), http.StatusPreconditionFailed)
		return
	}

	wc.UpdateStripeWebhook(ctx, webhookType, data)
}

func (wc *webhookController) UpdateStripeWebhook(ctx *gin.Context, webhookType string, body map[string]interface{}) {
	switch webhookType {
	case StripeWebhookCheckoutSessionCompleted, StripeWebhookInvoicePaid:
		wc.HandleStripeCheckoutAndInvoice(ctx, body, webhookType)
	default:
		apiresponse.NewSuccessResponse(ctx, fmt.Sprintf("webhook type %s not used", webhookType), http.StatusOK)
	}
}

func (wc *webhookController) HandleStripeCheckoutAndInvoice(ctx *gin.Context, body map[string]interface{}, webhookType string) {
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		apiresponse.NewErrorResponse(ctx, errors.New("data not found or invalid"), http.StatusPreconditionFailed)
		return
	}
	dataObject, ok := data["object"].(map[string]interface{})
	if !ok {
		apiresponse.NewErrorResponse(ctx, errors.New("data.object not found or invalid"), http.StatusPreconditionFailed)
		return
	}
	var checkoutSessionId, status, stripeSubscriptionId, latestInvoiceId, customerId string
	var filter bson.M
	var err error
	switch webhookType {
	case StripeWebhookCheckoutSessionCompleted:
		checkoutSessionId, ok = dataObject["id"].(string)
		if !ok {
			apiresponse.NewErrorResponse(ctx, errors.New("id not found in data object"), http.StatusPreconditionFailed)
			return
		}
		status, ok = dataObject["payment_status"].(string)
		if !ok {
			apiresponse.NewErrorResponse(ctx, errors.New("payment_status not found in data object"), http.StatusPreconditionFailed)
			return
		}
		latestInvoiceId, ok = dataObject["invoice"].(string)
		if !ok {
			fmt.Println("invoice not found in data object")
		}
		stripeSubscriptionId, ok = dataObject["subscription"].(string)
		if !ok {
			fmt.Println("subscription not found in data object")
		}
		customerId, ok = dataObject["customer"].(string)
		if !ok {
			fmt.Println("customer not found in data object")
		}
		filter = bson.M{"payment_details.payment_id": checkoutSessionId}
	case StripeWebhookInvoicePaid:
		status, stripeSubscriptionId, latestInvoiceId, customerId, err = extractStripeInvoiceFields(ctx, dataObject)
		if err != nil {
			apiresponse.NewErrorResponse(ctx, err, http.StatusPreconditionFailed)
			return
		}
		filter = bson.M{"payment_info.payment_id": stripeSubscriptionId}
	default:
		apiresponse.NewErrorResponse(ctx, errors.New("Unhandled webhook type - "+webhookType), http.StatusBadRequest)
		return
	}
	fmt.Println(status, latestInvoiceId, customerId, filter)
	// err  := wc.webhookService.UpdateBookingStatus()
}

func extractStripeInvoiceFields(ctx *gin.Context, dataObject map[string]interface{}) (status, stripeSubscriptionId, latestInvoiceId, customerId string, err error) {
	var ok bool
	status, ok = dataObject["status"].(string)
	if !ok {
		return status, stripeSubscriptionId, latestInvoiceId, customerId, fmt.Errorf("status not found in data object")
	}
	stripeSubscriptionId, ok = dataObject["subscription"].(string)
	if !ok {
		return status, stripeSubscriptionId, latestInvoiceId, customerId, fmt.Errorf("subscription not found in data object")
	}
	latestInvoiceId, ok = dataObject["id"].(string)
	if !ok {
		return status, stripeSubscriptionId, latestInvoiceId, customerId, fmt.Errorf("id not found in data object")
	}
	customerId, ok = dataObject["customer"].(string)
	if !ok {
		return status, stripeSubscriptionId, latestInvoiceId, customerId, fmt.Errorf("customer not found in data object")
	}
	return status, stripeSubscriptionId, latestInvoiceId, customerId, nil
}

func NewWebhookController() WebhookController {
	return &webhookController{}
}
