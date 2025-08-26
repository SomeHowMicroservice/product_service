package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type SizeRepository interface {
	Create(ctx context.Context, size *model.Size) error

	FindAll(ctx context.Context) ([]*model.Size, error)

	ExistsByID(ctx context.Context, id string) (bool, error)

	FindByID(ctx context.Context, id string) (*model.Size, error)

	FindByIDTx(ctx context.Context, tx *gorm.DB, id string) (*model.Size, error)

	Update(ctx context.Context, id string, updateData map[string]interface{}) error

	UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error

	FindAllByID(ctx context.Context, ids []string) ([]*model.Size, error)

	UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error

	FindAllDeleted(ctx context.Context) ([]*model.Size, error)

	FindDeletedByID(ctx context.Context, id string) (*model.Size, error)

	FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Size, error)

	DeleteAllByID(ctx context.Context, ids []string) error

	Delete(ctx context.Context, id string) error
}
