package models

type Booking struct {
	ID          string      `json:"id"`
	BookingId   string      `json:"booking_id"`
	UserId      string      `json:"user_id"`
	ShowId      string      `json:"show_id"`
	SeatNumbers []string    `json:"seat_numbers"`
	PaymentInfo PaymentInfo `json:"payment_info"`
	Status      string      `json:"status"`
}

type PaymentInfo struct {
	PaymentId       string  `json:"payment_id"`
	PaymentStatus   string  `json:"payment_status"`
	PaymentMethod   string  `json:"payment_method"`
	TransactionTime string  `json:"transaction_time"`
	PaymentAmount   float64 `json:"payment_amount"`
}

type InitiatePaymentResponse struct {
	PaymentLink string `json:"payment_link"`
	*Booking
}
