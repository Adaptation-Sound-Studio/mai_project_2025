package genre

import (
	"context"
	"upload-service/internal/domain/model"
)

type Repository interface {
	GetAllGenres(ctx context.Context) ([]model.Genre, error)
	CreateGenre(ctx context.Context, genre *model.Genre) (int64, error)
	UpdateGenre(ctx context.Context, genre *model.Genre) error
}
