package repository

import (
	"context"
	"time"

	"ticketBookingSystem/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type KnowledgeRepository interface {
	CreateKnowledgeChunk(ctx context.Context, chunk models.KnowledgeChunk) (models.KnowledgeChunk, error)
}

type knowledgeRepository struct {
	db *mongo.Database
}

func (kn *knowledgeRepository) CreateKnowledgeChunk(
	ctx context.Context,
	chunk models.KnowledgeChunk,
) (models.KnowledgeChunk, error) {

	collection := kn.db.Collection("knowledge_chunks")

	filter := bson.M{
		"source_id": chunk.SourceID,
	}

	update := bson.M{
		"$set": bson.M{
			"source_id":  chunk.SourceID,
			"content":    chunk.Content,
			"embedding":  chunk.Embedding,
			"metadata":   chunk.Metadata,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		ctx,
		filter,
		update,
		options.Update().SetUpsert(true),
	)

	if err != nil {
		return models.KnowledgeChunk{}, err
	}

	return chunk, nil
}

func NewKnowledgeRepository(client *mongo.Client) KnowledgeRepository {
	return &knowledgeRepository{db: client.Database("ticket-booking-system")}
}
