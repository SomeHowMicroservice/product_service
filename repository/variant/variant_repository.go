package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type VariantRepository interface {
	Create(ctx context.Context, variant *model.Variant) error

	CreateAllTx(ctx context.Context, tx *gorm.DB, variants []*model.Variant) error

	ExistsBySKU(ctx context.Context, sku string) (bool, error)

	FindAllByID(ctx context.Context, ids []string) ([]*model.Variant, error)

	UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error

	DeleteAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) error
}
