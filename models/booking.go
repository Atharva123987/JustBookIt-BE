package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Booking struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BookingId       string             `json:"booking_id" bson:"booking_id"`
	UserId          string             `json:"user_id" bson:"user_id"`
	ShowId          string             `json:"show_id" bson:"show_id"`
	MovieName       string             `json:"movie_name" bson:"movie_name"`
	ShowTimingEpoch int64              `json:"show_timing_epoch" bson:"show_timing_epoch"`
	MoviePoster     string             `json:"movie_poster" bson:"movie_poster"`
	SeatNumbers     []string           `json:"seat_numbers" bson:"seat_numbers"`
	PaymentInfo     PaymentInfo        `json:"payment_info" bson:"payment_info"`
	Status          string             `json:"status" bson:"status"`
}

type PaymentInfo struct {
	PaymentId       string  `json:"payment_id" bson:"payment_id"`
	PaymentStatus   string  `json:"payment_status" bson:"payment_status"`
	PaymentMethod   string  `json:"payment_method" bson:"payment_method"`
	TransactionTime string  `json:"transaction_time" bson:"transaction_time"`
	PaymentAmount   float64 `json:"payment_amount" bson:"payment_amount"`
}

type InitiatePaymentResponse struct {
	PaymentLink string `json:"payment_link" bson:"payment_link"`
	*Booking    `bson:",inline"`
}
