package service

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	productpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/product"
)

type ProductService interface {
	CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (string, error)

	GetCategoryTree(ctx context.Context) ([]*model.Category, error)

	GetCategoriesNoChild(ctx context.Context) ([]*model.Category, error)

	GetCategoriesNoProduct(ctx context.Context) ([]*model.Category, error)

	GetProductBySlug(ctx context.Context, productSlug string) (*model.Product, error)

	CreateColor(ctx context.Context, req *productpb.CreateColorRequest) (string, error)

	CreateSize(ctx context.Context, req *productpb.CreateSizeRequest) (string, error)

	GetProductsByCategory(ctx context.Context, categorySlug string) ([]*model.Product, error)

	CreateTag(ctx context.Context, req *productpb.CreateTagRequest) (string, error)

	GetAllCategories(ctx context.Context) ([]*model.Category, error)

	GetCategoryByID(ctx context.Context, categoryID string) (*productpb.CategoryAdminDetailsResponse, error)

	UpdateCategory(ctx context.Context, req *productpb.UpdateCategoryRequest) (*productpb.CategoryAdminDetailsResponse, error)

	GetAllColorsAdmin(ctx context.Context) (*productpb.ColorsAdminResponse, error)

	GetAllSizesAdmin(ctx context.Context) (*productpb.SizesAdminResponse, error)

	GetAllTagsAdmin(ctx context.Context) (*productpb.TagsAdminResponse, error)

	UpdateTag(ctx context.Context, req *productpb.UpdateTagRequest) error

	GetAllColors(ctx context.Context) ([]*model.Color, error)

	GetAllSizes(ctx context.Context) ([]*model.Size, error)

	GetAllTags(ctx context.Context) ([]*model.Tag, error)

	CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (string, error)

	GetProductByID(ctx context.Context, productID string) (*productpb.ProductAdminDetailsResponse, error)

	GetAllProductsAdmin(ctx context.Context, req *productpb.GetAllProductsAdminRequest) ([]*model.Product, *common.PaginationMeta, error)

	UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductAdminDetailsResponse, error)

	DeleteProduct(ctx context.Context, req *productpb.DeleteOneRequest) error

	DeleteProducts(ctx context.Context, req *productpb.DeleteManyRequest) error

	PermanentlyDeleteCategory(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error

	PermanentlyDeleteCategories(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error

	UpdateColor(ctx context.Context, req *productpb.UpdateColorRequest) error

	UpdateSize(ctx context.Context, req *productpb.UpdateSizeRequest) error

	DeleteColor(ctx context.Context, req *productpb.DeleteOneRequest) error

	DeleteSize(ctx context.Context, req *productpb.DeleteOneRequest) error

	DeleteColors(ctx context.Context, req *productpb.DeleteManyRequest) error

	DeleteSizes(ctx context.Context, req *productpb.DeleteManyRequest) error

	GetDeletedProducts(ctx context.Context, req *productpb.GetAllProductsAdminRequest) ([]*model.Product, *common.PaginationMeta, error)

	GetDeletedProductByID(ctx context.Context, productID string) (*productpb.ProductAdminDetailsResponse, error)

	GetDeletedColors(ctx context.Context) (*productpb.ColorsAdminResponse, error)

	GetDeletedSizes(ctx context.Context) (*productpb.SizesAdminResponse, error)

	GetDeletedTags(ctx context.Context) (*productpb.TagsAdminResponse, error)

	DeleteTag(ctx context.Context, req *productpb.DeleteOneRequest) error

	DeleteTags(ctx context.Context, req *productpb.DeleteManyRequest) error

	RestoreProduct(ctx context.Context, req *productpb.RestoreOneRequest) error

	RestoreProducts(ctx context.Context, req *productpb.RestoreManyRequest) error

	RestoreColor(ctx context.Context, req *productpb.RestoreOneRequest) error

	RestoreColors(ctx context.Context, req *productpb.RestoreManyRequest) error

	RestoreSize(ctx context.Context, req *productpb.RestoreOneRequest) error

	RestoreSizes(ctx context.Context, req *productpb.RestoreManyRequest) error

	RestoreTag(ctx context.Context, req *productpb.RestoreOneRequest) error

	RestoreTags(ctx context.Context, req *productpb.RestoreManyRequest) error

	PermanentlyDeleteProduct(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error

	PermanentlyDeleteProducts(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error

	PermanentlyDeleteColor(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error

	PermanentlyDeleteColors(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error

	PermanentlyDeleteSize(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error

	PermanentlyDeleteSizes(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error

	PermanentlyDeleteTag(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error

	PermanentlyDeleteTags(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error
}
