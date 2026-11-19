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
	VectorSearchChunks(ctx context.Context, params VectorSearchParams) ([]models.KnowledgeChunk, error)
}

// VectorSearchParams controls a MongoDB Atlas $vectorSearch over knowledge_chunks.
type VectorSearchParams struct {
	IndexName        string
	QueryVector      []float32
	Limit            int
	NumCandidates    int
	MinScore         float64
	ExcludeSourceIDs []string // chunks whose source_id is in this list are dropped
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

// VectorSearchChunks returns the knowledge chunks closest to the query vector,
// ordered by similarity (most similar first).
func (kn *knowledgeRepository) VectorSearchChunks(
	ctx context.Context,
	params VectorSearchParams,
) ([]models.KnowledgeChunk, error) {

	collection := kn.db.Collection("knowledge_chunks")

	// Ask for extra results so that excluding already-returned movies
	// still leaves up to params.Limit chunks.
	limit := params.Limit + len(params.ExcludeSourceIDs)
	numCandidates := params.NumCandidates
	if numCandidates < limit {
		numCandidates = limit
	}

	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: bson.M{
			"index":         params.IndexName,
			"path":          "embedding",
			"queryVector":   params.QueryVector,
			"numCandidates": numCandidates,
			"limit":         limit,
		}}},
		{{Key: "$project", Value: bson.M{
			"source_id": 1,
			"content":   1,
			"metadata":  1,
			"score":     bson.M{"$meta": "vectorSearchScore"},
		}}},
	}

	match := bson.M{"score": bson.M{"$gte": params.MinScore}}
	if len(params.ExcludeSourceIDs) > 0 {
		match["source_id"] = bson.M{"$nin": params.ExcludeSourceIDs}
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$limit", Value: params.Limit}},
	)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chunks []models.KnowledgeChunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}

func NewKnowledgeRepository(client *mongo.Client) KnowledgeRepository {
	return &knowledgeRepository{db: client.Database("ticket-booking-system")}
}
