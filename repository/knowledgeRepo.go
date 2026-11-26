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
	VectorSearch(ctx context.Context, embedding []float32, indexName string, limit int, minScore float64) ([]models.KnowledgeChunk, error)
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

// VectorSearch runs an Atlas $vectorSearch on knowledge_chunks.embedding and returns
// the matching chunks (best match first) whose similarity score is at least minScore.
// Only _id, source and source_id are populated on the returned chunks.
func (kn *knowledgeRepository) VectorSearch(
	ctx context.Context,
	embedding []float32,
	indexName string,
	limit int,
	minScore float64,
) ([]models.KnowledgeChunk, error) {
	if len(embedding) == 0 || limit < 1 {
		return []models.KnowledgeChunk{}, nil
	}

	numCandidates := limit * 20
	if numCandidates < 100 {
		numCandidates = 100
	}
	if numCandidates > 10000 {
		numCandidates = 10000
	}

	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: bson.M{
			"index":         indexName,
			"path":          "embedding",
			"queryVector":   embedding,
			"numCandidates": numCandidates,
			"limit":         limit,
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":       1,
			"source":    1,
			"source_id": 1,
			"score":     bson.M{"$meta": "vectorSearchScore"},
		}}},
		{{Key: "$match", Value: bson.M{"score": bson.M{"$gte": minScore}}}},
	}

	cursor, err := kn.db.Collection("knowledge_chunks").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := []models.KnowledgeChunk{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func NewKnowledgeRepository(client *mongo.Client) KnowledgeRepository {
	return &knowledgeRepository{db: client.Database("ticket-booking-system")}
}
