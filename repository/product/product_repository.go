package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error

	Delete(ctx context.Context, id string) error

	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	FindBySlugWithDetails(ctx context.Context, slug string) (*model.Product, error)

	FindByIDWithDetails(ctx context.Context, id string) (*model.Product, error)

	ExistsByID(ctx context.Context, id string) (bool, error)

	FindByID(ctx context.Context, id string) (*model.Product, error)

	FindByIDWithCategoriesAndTags(ctx context.Context, id string) (*model.Product, error)

	FindByIDWithCategoriesAndTagsTx(ctx context.Context, tx *gorm.DB, id string) (*model.Product, error)

	FindAllByCategorySlug(ctx context.Context, categorySlug string) ([]*model.Product, error)

	UpdateCategoriesTx(ctx context.Context, tx *gorm.DB, product *model.Product, categories []*model.Category) error

	UpdateTagsTx(ctx context.Context, tx *gorm.DB, product *model.Product, tags []*model.Tag) error

	Update(ctx context.Context, id string, updateData map[string]interface{}) error

	UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error

	FindAllByID(ctx context.Context, ids []string) ([]*model.Product, error)

	UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error

	FindAllDeletedPaginatedWithCategoriesAndThumbnail(ctx context.Context, query *common.PaginationQuery) ([]*model.Product, int64, error)

	FindDeletedByIDWithDetails(ctx context.Context, id string) (*model.Product, error)

	FindDeletedByID(ctx context.Context, id string) (*model.Product, error)

	FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Product, error)

	FindDeletedByIDWithImages(ctx context.Context, id string) (*model.Product, error)

	FindAllDeletedByIDWithImages(ctx context.Context, ids []string) ([]*model.Product, error)

	DeleteAllByID(ctx context.Context, ids []string) error

	FindAllPaginatedWithCategoriesAndThumbnail(ctx context.Context, query *common.PaginationQuery) ([]*model.Product, int64, error) 
}
