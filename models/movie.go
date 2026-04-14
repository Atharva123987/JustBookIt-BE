package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Movie struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title"`
	Duration    time.Duration      `json:"duration,omitempty" bson:"duration"`
	ReleaseDate time.Time          `json:"release_date,omitempty" bson:"release_date"`
	Genres      []string           `json:"genres,omitempty" bson:"genres"`
	Cast        []string           `json:"cast,omitempty" bson:"cast"`
	Description string             `json:"description,omitempty" bson:"description"`
	PosterURL   string             `json:"poster_url,omitempty" bson:"poster_url"`
}

type Show struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	MovieID         string             `json:"movie_id" bson:"movie_id"`
	MovieTitle      string             `json:"movie_title" bson:"movie_title"`
	ScreenNumber    int32              `json:"screen_number" bson:"screen_number"`
	ShowTimingEpoch int64              `json:"show_timing_epoch" bson:"show_timing_epoch"`
	ShowType        string             `json:"show_type" bson:"show_type"`
	LocationID      string             `json:"location_id" bson:"location_id"`
	Exits           []string           `json:"exits,omitempty" bson:"exits"`
	SeatRows        []SeatRow          `json:"seat_rows,omitempty" bson:"seat_rows"`
	PriceLabel      string             `json:"price_label,omitempty" bson:"price_label"`

	// ❗ Response-only field (not stored in DB)
	ShowTiming *time.Time `json:"show_timing,omitempty" bson:"-"`
}

type SeatRow struct {
	Row   string `json:"row" bson:"row"`
	Seats []Seat `json:"seats" bson:"seats"`
}

type Seat struct {
	Type       string  `json:"type" bson:"type"`     // VIP, Regular, AISLE
	Status     string  `json:"status" bson:"status"` // available, booked
	BookingID  string  `json:"booking_id,omitempty" bson:"booking_id,omitempty"`
	SeatNumber string  `json:"seat_number" bson:"seat_number"` // A1, B1, etc.
	Cost       float64 `json:"cost" bson:"cost"`               // price per seat
}

type Location struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Address  string             `json:"address" bson:"address"`
	Location MongoLocation      `json:"location" bson:"location"`
}

type MongoLocation struct {
	Type        string    `json:"type" bson:"type"`               // Point or Polygon
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [lng, lat]
}

type LocationMessage struct {
	Lat *float64 `json:"lat,omitempty" bson:"lat,omitempty"`
	Lng *float64 `json:"lng,omitempty" bson:"lng,omitempty"`
}

type LocationWithShow struct {
	Location `bson:",inline"`
	Shows    []map[string][]Show `json:"shows" bson:"shows"`
}
