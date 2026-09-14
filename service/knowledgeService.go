package service

import (
	"context"
	"ticketBookingSystem/models"
	"ticketBookingSystem/repository"
)

// ---- Interface ----

type KnowledgeService interface {
	CreateKnowledgeChunk(ctx context.Context, chunk models.KnowledgeChunk) (models.KnowledgeChunk, error)
}

// ---- Implementation ----

type knowledgeService struct {
	repo repository.KnowledgeRepository
}

// Constructor
func NewKnowledgeService(repo repository.KnowledgeRepository) KnowledgeService {
	return &knowledgeService{
		repo: repo,
	}
}

func (kn *knowledgeService) CreateKnowledgeChunk(ctx context.Context, chunk models.KnowledgeChunk) (models.KnowledgeChunk, error) {
	createdChunk, err := kn.repo.CreateKnowledgeChunk(ctx, chunk)
	return createdChunk, err
}
