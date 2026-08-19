package usecase

import (
	"context"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
)

type TrackUseCase struct {
	repo repository.TrackRepositoryInterface
}

func NewTrackUseCase(
	repo repository.TrackRepositoryInterface,
) *TrackUseCase {
	return &TrackUseCase{
		repo: repo,
	}
}

func (uc *TrackUseCase) CreateTrack(
	ctx context.Context,
	track *domain.Track,
) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	normalizeTrack(track)

	if err := validateTrack(track); err != nil {
		return err
	}

	return uc.repo.Create(ctx, track)
}

func (uc *TrackUseCase) GetTrackByID(
	ctx context.Context,
	id string,
) (*domain.Track, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, domain.ErrInvalidTrackID
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *TrackUseCase) UpdateTrack(
	ctx context.Context,
	track *domain.Track,
) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	normalizeTrack(track)

	if err := validateTrack(track); err != nil {
		return err
	}

	return uc.repo.Update(ctx, track)
}

func (uc *TrackUseCase) DeleteTrack(
	ctx context.Context,
	id string,
) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return domain.ErrInvalidTrackID
	}

	// Repository сам вернёт ErrTrackNotFound,
	// если объекта нет.
	return uc.repo.Delete(ctx, id)
}

func (uc *TrackUseCase) GetAllTracks(
	ctx context.Context,
) ([]*domain.Track, error) {
	return uc.repo.GetAllTracks(ctx)
}

func normalizeTrack(track *domain.Track) {
	track.ID = strings.TrimSpace(track.ID)
	track.Title = strings.TrimSpace(track.Title)
	track.Artist = strings.TrimSpace(track.Artist)
	track.URL = strings.TrimSpace(track.URL)
	track.AlbumTitle = strings.TrimSpace(track.AlbumTitle)
	track.AlbumArtURL = strings.TrimSpace(track.AlbumArtURL)
	track.PreviewURL = strings.TrimSpace(track.PreviewURL)

	for i := range track.Genre {
		track.Genre[i] = strings.TrimSpace(track.Genre[i])
	}
}

func validateTrack(track *domain.Track) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	if track.ID == "" {
		return domain.ErrInvalidTrackID
	}

	if track.Title == "" {
		return domain.ErrInvalidTrackTitle
	}

	if track.Artist == "" {
		return domain.ErrInvalidTrackArtist
	}

	if track.URL == "" {
		return domain.ErrInvalidTrackURL
	}

	if track.Year < 1800 || track.Year > 2100 {
		return domain.ErrInvalidTrackYear
	}

	return nil
}
