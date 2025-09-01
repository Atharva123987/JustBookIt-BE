package service

import (
	"ticketBookingSystem/models"
)

type PaymentService interface {
	InitiatePayment(amount float64, customerID string) (*models.InitiatePaymentResponse, error)
}

type paymentService struct{}

func NewPaymentService() PaymentService {
	return &paymentService{}
}

func (ps *paymentService) InitiatePayment(amount float64, customerID string) (*models.InitiatePaymentResponse, error) {
	gateway := GetPaymentGatewayService("stripe") // Replace with actual payment gateway implementation
	res, err := gateway.InitiatePayment(int64(amount))
	if err != nil {
		return nil, err
	}
	return res, nil
}
