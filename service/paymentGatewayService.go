package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ticketBookingSystem/internal/config"
	"ticketBookingSystem/models"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

type PaymentGatewayService interface {
	InitiatePayment(amount int64, bookingId string) (*models.InitiatePaymentResponse, error)
}

type Stripe struct{}

var paymentGatewayMap = map[string]PaymentGatewayService{
	"stripe": &Stripe{},
}

func GetPaymentGatewayService(name string) PaymentGatewayService {
	return paymentGatewayMap[name]
}

func ExecuteApiRequest(req *http.Request) (map[string]any, error) {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request %v", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-ok status: %v, body: %v", res.StatusCode, string(body))
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response %v", err)
	}
	return apiResponse, nil
}

func (s *Stripe) InitiatePayment(amount int64, bookingId string) (*models.InitiatePaymentResponse, error) {
	stripe.Key = config.AppConfig.StripeSecret // your real secret key here

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String("payment"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("inr"),
					UnitAmount: stripe.Int64(amount * 100), // $10.00
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Movie Ticket"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		ClientReferenceID: stripe.String(bookingId),
		SuccessURL:        stripe.String("ticketbookingsystemfe://payment/success?bookingId=" + bookingId),
		CancelURL:         stripe.String("ticketbookingsystemfe://payment/cancel?bookingId=" + bookingId),
	}

	// This will auto-create a Customer if you don't set params.Customer
	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}
	return &models.InitiatePaymentResponse{
		PaymentLink: sess.URL,
		Booking: &models.Booking{
			PaymentInfo: models.PaymentInfo{
				PaymentAmount: float64(amount),
			},
		},
	}, nil
}
