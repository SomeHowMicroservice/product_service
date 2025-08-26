package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error

	FindByID(ctx context.Context, id string) (*model.Category, error)

	FindAllByID(ctx context.Context, ids []string) ([]*model.Category, error)

	FindAllByIDWithChildren(ctx context.Context, ids []string) ([]*model.Category, error)

	FindAllWithProducts(ctx context.Context) ([]*model.Category, error)

	ExistsByID(ctx context.Context, id string) (bool, error)

	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	ExistsBySlugTx(ctx context.Context, tx *gorm.DB, slug string) (bool, error)

	FindAllWithParentsAndChildren(ctx context.Context) ([]*model.Category, error)

	FindAll(ctx context.Context) ([]*model.Category, error)

	FindByIDWithParentsAndProducts(ctx context.Context, id string) (*model.Category, error)

	UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error

	FindByIDWithParents(ctx context.Context, id string) (*model.Category, error)

	UpdateParentsTx(ctx context.Context, tx *gorm.DB, category *model.Category, parents []*model.Category) error

	FindAllWithChildren(ctx context.Context) ([]*model.Category, error)

	DeleteAllByID(ctx context.Context, ids []string) error

	Delete(ctx context.Context, id string) error

	FindByIDWithParentsTx(ctx context.Context, tx *gorm.DB, id string) (*model.Category, error)

	FindAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) ([]*model.Category, error)

	GetAllAncestors(ctx context.Context, id string) ([]string, error)

	GetAllDescendants(ctx context.Context, id string) ([]string, error)
}