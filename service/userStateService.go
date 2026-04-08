package service

import (
	"context"
	"ticketBookingSystem/repository"
)

// ---- Interface ----

type UserStateService interface {
	GetState(ctx context.Context, userID string) (*repository.UserState, error)
	UpdateState(ctx context.Context, userID string, newState repository.UserState) error
	ClearIntent(ctx context.Context, userID string) error
}

// ---- Implementation ----

type userStateService struct {
	repo repository.UserStateRepository
}

// Constructor
func NewUserStateService(repo repository.UserStateRepository) UserStateService {
	return &userStateService{
		repo: repo,
	}
}

// ---- Methods ----

// Get full state
func (s *userStateService) GetState(ctx context.Context, userID string) (*repository.UserState, error) {
	return s.repo.GetState(ctx, userID)
}

func (s *userStateService) UpdateState(ctx context.Context, userID string, newState repository.UserState) error {
	current, err := s.repo.GetState(ctx, userID)
	if err != nil {
		// If not found, just create
		return s.repo.UpdateState(ctx, userID, newState)
	}

	// Merge fields (only overwrite if not nil)
	if newState.Movie != nil {
		current.Movie = newState.Movie
	}

	if newState.NoOfSeats != nil {
		current.NoOfSeats = newState.NoOfSeats
	}
	if newState.SeatNumbers != nil {
		current.SeatNumbers = newState.SeatNumbers
	}
	if newState.MovieID != nil {
		current.MovieID = newState.MovieID
	}
	if newState.Intent != nil {
		current.Intent = newState.Intent
	}
	if newState.ShowID != nil {
		current.ShowID = newState.ShowID
	}

	return s.repo.UpdateState(ctx, userID, *current)
}

// Clear intent
func (s *userStateService) ClearIntent(ctx context.Context, userID string) error {
	return s.repo.ClearIntent(ctx, userID)
}
