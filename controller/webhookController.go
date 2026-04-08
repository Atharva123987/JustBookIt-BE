package controller

import (
	"errors"
	"fmt"
	"net/http"
	apiresponse "ticketBookingSystem/internal/apiResponse"
	"ticketBookingSystem/service"

	"github.com/gin-gonic/gin"
)

type WebhookController interface {
	StripeWebhook(ctx *gin.Context)
}

var StripeWebhookCheckoutSessionCompleted string = "checkout.session.completed"
var StripeWebhookInvoicePaid = "invoice.paid"

type webhookController struct {
	bookingService service.BookingService
}

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
	var status, clientReferenceId, customerId string
	switch webhookType {
	case StripeWebhookCheckoutSessionCompleted:

		status, ok = dataObject["payment_status"].(string)
		if !ok {
			apiresponse.NewErrorResponse(ctx, errors.New("payment_status not found in data object"), http.StatusPreconditionFailed)
			return
		}
		fmt.Print(status)
		customerId, ok = dataObject["customer"].(string)
		if !ok {
			fmt.Println("customer not found in data object")
		}
		clientReferenceId, ok = dataObject["client_reference_id"].(string)
		if !ok {
			fmt.Println("client_reference_id not found in data object")
		}

	default:
		apiresponse.NewErrorResponse(ctx, errors.New("Unhandled webhook type - "+webhookType), http.StatusBadRequest)
		return
	}
	err := wc.bookingService.ConfirmBooking(ctx, customerId, clientReferenceId)
	if err != nil {
		fmt.Printf("error while confirming booking %s", err.Error())
		return
	}
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

func NewWebhookController(bookingService service.BookingService) WebhookController {
	return &webhookController{
		bookingService: bookingService,
	}

}
