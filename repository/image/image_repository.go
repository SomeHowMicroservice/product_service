package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type ImageRepository interface {
	Create(ctx context.Context, image *model.Image) error

	CreateAllTx(ctx context.Context, tx *gorm.DB, images []*model.Image) error

	FindAllByID(ctx context.Context, ids []string) ([]*model.Image, error)

	Update(ctx context.Context, id string, updateData map[string]interface{}) error

	UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error

	DeleteAllByID(ctx context.Context, ids []string) error

	DeleteAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) error

	UpdateFileID(ctx context.Context, id string, fileID string) error
}
