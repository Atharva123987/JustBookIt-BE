package service

import (
	"context"
	"errors"
	"ticketBookingSystem/repository"
)

// ---- Interface ----

type UserStateService interface {
	GetState(ctx context.Context, deviceID string) (*repository.UserState, error)
	UpdateState(ctx context.Context, deviceID string, newState repository.UserState) error
	ClearState(ctx context.Context, deviceID string) error
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
func (s *userStateService) GetState(ctx context.Context, deviceID string) (*repository.UserState, error) {
	if deviceID == "" {
		return nil, errors.New("device id is empty")
	}
	return s.repo.GetState(ctx, deviceID)
}

func (s *userStateService) UpdateState(ctx context.Context, deviceID string, newState repository.UserState) error {
	if deviceID == "" {
		return errors.New("device id is empty")
	}
	current, err := s.repo.GetState(ctx, deviceID)
	if err != nil {
		// If not found, just create
		return s.repo.UpdateState(ctx, deviceID, newState)
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

	return s.repo.UpdateState(ctx, deviceID, *current)
}

// Clear intent
func (s *userStateService) ClearState(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return errors.New("device id is empty")
	}
	return s.repo.ClearState(ctx, deviceID)
}
