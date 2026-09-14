package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KnowledgeChunk struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Source    string             `bson:"source" json:"source"`
	SourceID  string             `bson:"source_id" json:"source_id"`
	Content   string             `bson:"content" json:"content"`
	Embedding []float32          `bson:"embedding" json:"embedding"`
	Metadata  map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
