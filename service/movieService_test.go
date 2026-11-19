package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"ticketBookingSystem/internal/config"
	"ticketBookingSystem/models"
	"ticketBookingSystem/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type fakeMovieRepo struct {
	repository.MovieRepository
	textResults []models.Movie
	all         map[primitive.ObjectID]models.Movie
}

func (f *fakeMovieRepo) GetAllMovies(_ context.Context, filter bson.M, _ ...*options.FindOptions) ([]models.Movie, error) {
	idFilter, _ := filter["_id"].(bson.M)
	if in, ok := idFilter["$in"]; ok {
		ids := in.([]primitive.ObjectID)
		// Return in reverse to prove the service restores similarity order.
		out := []models.Movie{}
		for i := len(ids) - 1; i >= 0; i-- {
			if m, ok := f.all[ids[i]]; ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	return f.textResults, nil
}

type fakeKnowledgeRepo struct {
	repository.KnowledgeRepository
	chunks     []models.KnowledgeChunk
	err        error
	lastParams repository.VectorSearchParams
	calls      int
}

func (f *fakeKnowledgeRepo) VectorSearchChunks(_ context.Context, p repository.VectorSearchParams) ([]models.KnowledgeChunk, error) {
	f.calls++
	f.lastParams = p
	if f.err != nil {
		return nil, f.err
	}
	// Mimic the $match on source_id $nin.
	out := []models.KnowledgeChunk{}
	for _, c := range f.chunks {
		excluded := false
		for _, id := range p.ExcludeSourceIDs {
			excluded = excluded || id == c.SourceID
		}
		if !excluded {
			out = append(out, c)
		}
	}
	return out, nil
}

func setup(t *testing.T) (*gin.Context, []models.Movie) {
	config.AppConfig = &config.Config{VectorSearchIndex: "idx", VectorSearchLimit: 5, VectorSearchNumCandidates: 100, VectorSearchMinScore: 0.7}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	movies := make([]models.Movie, 4)
	for i := range movies {
		movies[i] = models.Movie{ID: primitive.NewObjectID(), Title: string(rune('A' + i))}
	}
	return ctx, movies
}

func TestTextResultsFirstThenVectorResultsDeduped(t *testing.T) {
	ctx, m := setup(t)
	all := map[primitive.ObjectID]models.Movie{}
	for _, x := range m {
		all[x.ID] = x
	}
	mr := &fakeMovieRepo{textResults: []models.Movie{m[0]}, all: all}
	kr := &fakeKnowledgeRepo{chunks: []models.KnowledgeChunk{
		{SourceID: m[2].ID.Hex()},
		{SourceID: m[0].ID.Hex()}, // already a text result
		{SourceID: m[2].ID.Hex()}, // duplicate chunk for same movie
		{SourceID: m[1].ID.Hex()},
		{SourceID: "not-an-object-id"},
	}}
	svc := NewMovieService(mr, kr)

	got, cursor, err := svc.GetAllMovies(ctx, 10, "", "a", []float32{0.1, 0.2})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A", "C", "B"}
	if len(got) != len(want) {
		t.Fatalf("got %d movies, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Title != w {
			t.Errorf("position %d: got %s, want %s", i, got[i].Title, w)
		}
	}
	if cursor != m[0].ID.Hex() {
		t.Errorf("cursor should follow text results, got %s", cursor)
	}
	if len(kr.lastParams.ExcludeSourceIDs) != 1 || kr.lastParams.ExcludeSourceIDs[0] != m[0].ID.Hex() {
		t.Errorf("text result ids should be excluded from vector search: %v", kr.lastParams.ExcludeSourceIDs)
	}
}

func TestNoEmbeddingsOrLaterPageSkipsVectorSearch(t *testing.T) {
	ctx, m := setup(t)
	mr := &fakeMovieRepo{textResults: []models.Movie{m[0]}}
	kr := &fakeKnowledgeRepo{}
	svc := NewMovieService(mr, kr)

	if _, _, err := svc.GetAllMovies(ctx, 10, "", "a", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.GetAllMovies(ctx, 10, m[3].ID.Hex(), "a", []float32{1}); err != nil {
		t.Fatal(err)
	}
	if kr.calls != 0 {
		t.Errorf("vector search should not run, ran %d times", kr.calls)
	}
}

func TestVectorSearchErrorFallsBackToTextResults(t *testing.T) {
	ctx, m := setup(t)
	mr := &fakeMovieRepo{textResults: []models.Movie{m[0], m[1]}}
	kr := &fakeKnowledgeRepo{err: errors.New("index not found")}
	svc := NewMovieService(mr, kr)

	got, _, err := svc.GetAllMovies(ctx, 10, "", "a", []float32{1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("want 2 text results, got %d", len(got))
	}
}
