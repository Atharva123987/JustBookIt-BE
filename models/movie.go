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
	ShowId          string             `json:"show_id,omitempty" bson:"show_id"` // Unique identifier for the show (movie_id + show_type + location_id + screen_number)
	MovieID         string             `json:"movie_id" bson:"movie_id"`
	MovieTitle      string             `json:"movie_title" bson:"movie_title"`
	ScreenNumber    int32              `json:"screen_number" bson:"screen_number"`
	ShowTimingEpoch int64              `json:"show_timing_epoch" bson:"show_timing_epoch"`
	ShowTiming      time.Time          `json:"show_timing" bson:"show_timing"`
	ShowType        string             `json:"show_type" bson:"show_type"`
	LocationID      string             `json:"location_id" bson:"location_id"`
}

type Location struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Address  string             `json:"address" bson:"address"`
	Location MongoLocation      `json:"location" bson:"location"`
}

type MongoLocation struct {
	Type        string    `json:"type"` // Point or Polygon
	Coordinates []float64 `json:"coordinates"`
}

type LocationMessage struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type LocationWithShow struct {
	Location
	Shows []map[string][]Show `json:"shows"`
}
