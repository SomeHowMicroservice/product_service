package handler

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	productpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/product"
	"github.com/SomeHowMicroservice/shm-be/product/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type GRPCHandler struct {
	productpb.UnimplementedProductServiceServer
	svc service.ProductService
}

func NewGRPCHandler(grpcServer *grpc.Server, svc service.ProductService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (*productpb.CreatedResponse, error) {
	categoryID, err := h.svc.CreateCategory(ctx, req)
	if err != nil {
		switch err {
		case common.ErrSlugAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case common.ErrCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.CreatedResponse{
		Id: categoryID,
	}, nil
}

func (h *GRPCHandler) GetCategoryTree(ctx context.Context, req *productpb.GetManyRequest) (*productpb.CategoryTreeResponse, error) {
	categoryTree, err := h.svc.GetCategoryTree(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toCategoryTreeResponse(categoryTree), nil
}

func (h *GRPCHandler) GetCategoriesNoProduct(ctx context.Context, req *productpb.GetManyRequest) (*productpb.BaseCategoriesResponse, error) {
	categories, err := h.svc.GetCategoriesNoProduct(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &productpb.BaseCategoriesResponse{
		Categories: toBaseCategoriesResponse(categories),
	}, nil
}

func (h *GRPCHandler) GetProductBySlug(ctx context.Context, req *productpb.GetProductBySlugRequest) (*productpb.ProductPublicResponse, error) {
	product, err := h.svc.GetProductBySlug(ctx, req.Slug)
	if err != nil {
		switch err {
		case common.ErrProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toProductPublicResponse(product), nil
}

func (h *GRPCHandler) CreateColor(ctx context.Context, req *productpb.CreateColorRequest) (*productpb.CreatedResponse, error) {
	colorID, err := h.svc.CreateColor(ctx, req)
	if err != nil {
		switch err {
		case common.ErrColorAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.CreatedResponse{
		Id: colorID,
	}, nil
}

func (h *GRPCHandler) CreateSize(ctx context.Context, req *productpb.CreateSizeRequest) (*productpb.CreatedResponse, error) {
	sizeID, err := h.svc.CreateSize(ctx, req)
	if err != nil {
		switch err {
		case common.ErrSizeAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.CreatedResponse{
		Id: sizeID,
	}, nil
}

func (h *GRPCHandler) GetProductsByCategory(ctx context.Context, req *productpb.GetProductsByCategoryRequest) (*productpb.ProductsPublicResponse, error) {
	products, err := h.svc.GetProductsByCategory(ctx, req.Slug)
	if err != nil {
		switch err {
		case common.ErrCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return toProductsPublicResponse(products), nil
}

func (h *GRPCHandler) CreateTag(ctx context.Context, req *productpb.CreateTagRequest) (*productpb.CreatedResponse, error) {
	tagID, err := h.svc.CreateTag(ctx, req)
	if err != nil {
		switch err {
		case common.ErrTagAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.CreatedResponse{
		Id: tagID,
	}, nil
}

func (h *GRPCHandler) GetAllCategoriesAdmin(ctx context.Context, req *productpb.GetManyRequest) (*productpb.BaseCategoriesResponse, error) {
	categories, err := h.svc.GetAllCategories(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.BaseCategoriesResponse{
		Categories: toBaseCategoriesResponse(categories),
	}, nil
}

func (h *GRPCHandler) GetCategoryById(ctx context.Context, req *productpb.GetCategoryByIdRequest) (*productpb.CategoryAdminDetailsResponse, error) {
	convertedCategory, err := h.svc.GetCategoryByID(ctx, req.Id)
	if err != nil {
		switch err {
		case common.ErrUserNotFound, common.ErrCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedCategory, nil
}

func (h *GRPCHandler) UpdateCategory(ctx context.Context, req *productpb.UpdateCategoryRequest) (*productpb.CategoryAdminDetailsResponse, error) {
	convertedCategory, err := h.svc.UpdateCategory(ctx, req)
	if err != nil {
		switch err {
		case common.ErrUserNotFound, common.ErrHasCategoryNotFound, common.ErrCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedCategory, nil
}

func (h *GRPCHandler) GetAllColorsAdmin(ctx context.Context, req *productpb.GetManyRequest) (*productpb.ColorsAdminResponse, error) {
	convertedColors, err := h.svc.GetAllColorsAdmin(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedColors, nil
}

func (h *GRPCHandler) GetAllColors(ctx context.Context, req *productpb.GetManyRequest) (*productpb.ColorsPublicResponse, error) {
	colors, err := h.svc.GetAllColors(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toColorsPublicResponse(colors), nil
}

func (h *GRPCHandler) GetAllSizesAdmin(ctx context.Context, req *productpb.GetManyRequest) (*productpb.SizesAdminResponse, error) {
	convertedSizes, err := h.svc.GetAllSizesAdmin(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedSizes, nil
}

func (h *GRPCHandler) GetAllSizes(ctx context.Context, req *productpb.GetManyRequest) (*productpb.SizesPublicResponse, error) {
	sizes, err := h.svc.GetAllSizes(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toSizesPublicResponse(sizes), nil
}

func (h *GRPCHandler) GetAllTagsAdmin(ctx context.Context, req *productpb.GetManyRequest) (*productpb.TagsAdminResponse, error) {
	convertedTags, err := h.svc.GetAllTagsAdmin(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedTags, nil
}

func (h *GRPCHandler) GetAllTags(ctx context.Context, req *productpb.GetManyRequest) (*productpb.TagsPublicResponse, error) {
	tags, err := h.svc.GetAllTags(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var baseTags []*productpb.BaseTagResponse
	for _, tag := range tags {
		baseTags = append(baseTags, &productpb.BaseTagResponse{
			Id:   tag.ID,
			Name: tag.Name,
		})
	}

	return &productpb.TagsPublicResponse{
		Tags: baseTags,
	}, nil
}

func (h *GRPCHandler) UpdateTag(ctx context.Context, req *productpb.UpdateTagRequest) (*productpb.UpdatedResponse, error) {
	if err := h.svc.UpdateTag(ctx, req); err != nil {
		switch err {
		case common.ErrTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case common.ErrTagAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.UpdatedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreatedResponse, error) {
	productID, err := h.svc.CreateProduct(ctx, req)
	if err != nil {
		switch err {
		case common.ErrSlugAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case common.ErrHasCategoryNotFound, common.ErrHasTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.CreatedResponse{
		Id: productID,
	}, nil
}

func (h *GRPCHandler) GetCategoriesNoChild(ctx context.Context, req *productpb.GetManyRequest) (*productpb.BaseCategoriesResponse, error) {
	categories, err := h.svc.GetCategoriesNoChild(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.BaseCategoriesResponse{
		Categories: toBaseCategoriesResponse(categories),
	}, nil
}

func (h *GRPCHandler) GetProductById(ctx context.Context, req *productpb.GetProductByIdRequest) (*productpb.ProductAdminDetailsResponse, error) {
	convertedProduct, err := h.svc.GetProductByID(ctx, req.Id)
	if err != nil {
		switch err {
		case common.ErrProductNotFound, common.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedProduct, nil
}

func (h *GRPCHandler) GetAllProductsAdmin(ctx context.Context, req *productpb.GetAllProductsAdminRequest) (*productpb.ProductsAdminResponse, error) {
	products, meta, err := h.svc.GetAllProductsAdmin(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProductsAdminResponse(products, meta), nil
}

func (h *GRPCHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductAdminDetailsResponse, error) {
	convertedProduct, err := h.svc.UpdateProduct(ctx, req)
	if err != nil {
		switch err {
		case common.ErrSlugAlreadyExists, common.ErrHasSKUAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case common.ErrUserNotFound, common.ErrHasCategoryNotFound, common.ErrHasTagNotFound, common.ErrHasImageNotFound, common.ErrHasVariantNotFound, common.ErrProductNotFound, common.ErrVariantNotFound, common.ErrImageNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedProduct, nil
}

func (h *GRPCHandler) DeleteProduct(ctx context.Context, req *productpb.DeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteProduct(ctx, req); err != nil {
		switch err {
		case common.ErrProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteProducts(ctx context.Context, req *productpb.DeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteProducts(ctx, req); err != nil {
		switch err {
		case common.ErrHasProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteCategory(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteCategory(ctx, req); err != nil {
		switch err {
		case common.ErrCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteCategories(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteCategories(ctx, req); err != nil {
		switch err {
		case common.ErrHasCategoryNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) UpdateColor(ctx context.Context, req *productpb.UpdateColorRequest) (*productpb.UpdatedResponse, error) {
	if err := h.svc.UpdateColor(ctx, req); err != nil {
		switch err {
		case common.ErrColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case common.ErrColorAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.UpdatedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) UpdateSize(ctx context.Context, req *productpb.UpdateSizeRequest) (*productpb.UpdatedResponse, error) {
	if err := h.svc.UpdateSize(ctx, req); err != nil {
		switch err {
		case common.ErrSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case common.ErrSizeAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.UpdatedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteColor(ctx context.Context, req *productpb.DeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteColor(ctx, req); err != nil {
		switch err {
		case common.ErrColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteSize(ctx context.Context, req *productpb.DeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteSize(ctx, req); err != nil {
		switch err {
		case common.ErrSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteColors(ctx context.Context, req *productpb.DeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteColors(ctx, req); err != nil {
		switch err {
		case common.ErrHasColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteSizes(ctx context.Context, req *productpb.DeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteSizes(ctx, req); err != nil {
		switch err {
		case common.ErrHasSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) GetDeletedProducts(ctx context.Context, req *productpb.GetAllProductsAdminRequest) (*productpb.ProductsAdminResponse, error) {
	products, meta, err := h.svc.GetDeletedProducts(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProductsAdminResponse(products, meta), nil
}

func (h *GRPCHandler) GetDeletedProductById(ctx context.Context, req *productpb.GetProductByIdRequest) (*productpb.ProductAdminDetailsResponse, error) {
	convertedProduct, err := h.svc.GetDeletedProductByID(ctx, req.Id)
	if err != nil {
		switch err {
		case common.ErrProductNotFound, common.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedProduct, nil
}

func (h *GRPCHandler) GetDeletedColors(ctx context.Context, req *productpb.GetManyRequest) (*productpb.ColorsAdminResponse, error) {
	convertedColors, err := h.svc.GetDeletedColors(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedColors, nil
}

func (h *GRPCHandler) GetDeletedSizes(ctx context.Context, req *productpb.GetManyRequest) (*productpb.SizesAdminResponse, error) {
	convertedSizes, err := h.svc.GetDeletedSizes(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedSizes, nil
}

func (h *GRPCHandler) GetDeletedTags(ctx context.Context, req *productpb.GetManyRequest) (*productpb.TagsAdminResponse, error) {
	convertedTags, err := h.svc.GetDeletedTags(ctx)
	if err != nil {
		switch err {
		case common.ErrHasUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return convertedTags, nil
}

func (h *GRPCHandler) DeleteTag(ctx context.Context, req *productpb.DeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteTag(ctx, req); err != nil {
		switch err {
		case common.ErrTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) DeleteTags(ctx context.Context, req *productpb.DeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.DeleteTags(ctx, req); err != nil {
		switch err {
		case common.ErrHasTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreProduct(ctx context.Context, req *productpb.RestoreOneRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreProduct(ctx, req); err != nil {
		switch err {
		case common.ErrProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreProducts(ctx context.Context, req *productpb.RestoreManyRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreProducts(ctx, req); err != nil {
		switch err {
		case common.ErrHasProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreColor(ctx context.Context, req *productpb.RestoreOneRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreColor(ctx, req); err != nil {
		switch err {
		case common.ErrColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreColors(ctx context.Context, req *productpb.RestoreManyRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreColors(ctx, req); err != nil {
		switch err {
		case common.ErrHasColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreSize(ctx context.Context, req *productpb.RestoreOneRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreSize(ctx, req); err != nil {
		switch err {
		case common.ErrSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreSizes(ctx context.Context, req *productpb.RestoreManyRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreSizes(ctx, req); err != nil {
		switch err {
		case common.ErrHasSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreTag(ctx context.Context, req *productpb.RestoreOneRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreTag(ctx, req); err != nil {
		switch err {
		case common.ErrTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) RestoreTags(ctx context.Context, req *productpb.RestoreManyRequest) (*productpb.RestoredResponse, error) {
	if err := h.svc.RestoreTags(ctx, req); err != nil {
		switch err {
		case common.ErrHasTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.RestoredResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteProduct(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteProduct(ctx, req); err != nil {
		switch err {
		case common.ErrProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteProducts(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteProducts(ctx, req); err != nil {
		switch err {
		case common.ErrHasProductNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteColor(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteColor(ctx, req); err != nil {
		switch err {
		case common.ErrColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteColors(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteColors(ctx, req); err != nil {
		switch err {
		case common.ErrHasColorNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteSize(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteSize(ctx, req); err != nil {
		switch err {
		case common.ErrSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteSizes(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteSizes(ctx, req); err != nil {
		switch err {
		case common.ErrHasSizeNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteTag(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteTag(ctx, req); err != nil {
		switch err {
		case common.ErrTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func (h *GRPCHandler) PermanentlyDeleteTags(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) (*productpb.DeletedResponse, error) {
	if err := h.svc.PermanentlyDeleteTags(ctx, req); err != nil {
		switch err {
		case common.ErrHasTagNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &productpb.DeletedResponse{
		Success: true,
	}, nil
}

func toProductsAdminResponse(products []*model.Product, meta *common.PaginationMeta) *productpb.ProductsAdminResponse {
	var productResponses []*productpb.ProductAdminResponse
	for _, pro := range products {
		productResponses = append(productResponses, toProductAdminResponse(pro))
	}

	return &productpb.ProductsAdminResponse{
		Products: productResponses,
		Meta: &productpb.PaginationMetaResponse{
			Page:       uint32(meta.Page),
			Limit:      uint32(meta.Limit),
			Total:      uint32(meta.Total),
			TotalPages: uint32(meta.TotalPages),
			HasPrev:    &meta.HasPrev,
			HasNext:    &meta.HasNext,
		},
	}
}

func toProductAdminResponse(product *model.Product) *productpb.ProductAdminResponse {
	categories := toBaseCategoriesResponse(product.Categories)
	thumbnail := &productpb.SimpleImageResponse{
		Id:  product.Images[0].ID,
		Url: product.Images[0].Url,
	}
	return &productpb.ProductAdminResponse{
		Id:         product.ID,
		Title:      product.Title,
		Price:      product.Price,
		Categories: categories,
		Thumbnail:  thumbnail,
	}
}

func toProductsPublicResponse(products []*model.Product) *productpb.ProductsPublicResponse {
	var productResponses []*productpb.ProductPublicResponse
	for _, pro := range products {
		productResponses = append(productResponses, toProductPublicResponse(pro))
	}

	return &productpb.ProductsPublicResponse{
		Products: productResponses,
	}
}

func toProductPublicResponse(product *model.Product) *productpb.ProductPublicResponse {
	var startSalePtr, endSalePtr *string
	if product.StartSale != nil {
		formatted := product.StartSale.Format("2006-01-02")
		startSalePtr = &formatted
	}
	if product.EndSale != nil {
		formatted := product.EndSale.Format("2006-01-02")
		endSalePtr = &formatted
	}

	categories := make([]*productpb.BaseCategoryResponse, len(product.Categories))
	for i, category := range product.Categories {
		categories[i] = toBaseCategoryResponse(category)
	}

	variants := make([]*productpb.BaseVariantResponse, len(product.Variants))
	for i, variant := range product.Variants {
		variants[i] = toBaseVariantResponse(variant)
	}

	images := make([]*productpb.BaseImageResponse, len(product.Images))
	for i, image := range product.Images {
		images[i] = toBaseImageResponse(image)
	}

	return &productpb.ProductPublicResponse{
		Id:          product.ID,
		Title:       product.Title,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		IsSale:      &product.IsSale,
		SalePrice:   product.SalePrice,
		StartSale:   startSalePtr,
		EndSale:     endSalePtr,
		Categories:  categories,
		Variants:    variants,
		Images:      images,
	}
}

func toCategoryPublicResponse(category *model.Category) *productpb.CategoryPublicResponse {
	children := make([]*productpb.CategoryPublicResponse, 0, len(category.Children))
	for _, child := range category.Children {
		children = append(children, toCategoryPublicResponse(child))
	}

	return &productpb.CategoryPublicResponse{
		Id:       category.ID,
		Name:     category.Name,
		Slug:     category.Slug,
		Children: children,
	}
}

func toCategoryTreeResponse(categories []*model.Category) *productpb.CategoryTreeResponse {
	result := make([]*productpb.CategoryPublicResponse, 0, len(categories))
	for _, c := range categories {
		result = append(result, toCategoryPublicResponse(c))
	}

	return &productpb.CategoryTreeResponse{
		Categories: result,
	}
}

func toBaseCategoriesResponse(categories []*model.Category) []*productpb.BaseCategoryResponse {
	var baseCategories []*productpb.BaseCategoryResponse
	for _, category := range categories {
		baseCategories = append(baseCategories, toBaseCategoryResponse(category))
	}
	return baseCategories
}

func toBaseCategoryResponse(category *model.Category) *productpb.BaseCategoryResponse {
	return &productpb.BaseCategoryResponse{
		Id:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}
}

func toColorsPublicResponse(colors []*model.Color) *productpb.ColorsPublicResponse {
	var baseColors []*productpb.BaseColorResponse
	for _, color := range colors {
		baseColors = append(baseColors, toBaseColorResponse(color))
	}
	return &productpb.ColorsPublicResponse{
		Colors: baseColors,
	}
}

func toBaseColorResponse(color *model.Color) *productpb.BaseColorResponse {
	return &productpb.BaseColorResponse{
		Id:   color.ID,
		Name: color.Name,
	}
}

func toSizesPublicResponse(sizes []*model.Size) *productpb.SizesPublicResponse {
	var baseSizes []*productpb.BaseSizeResponse
	for _, size := range sizes {
		baseSizes = append(baseSizes, toBaseSizeResponse(size))
	}
	return &productpb.SizesPublicResponse{
		Sizes: baseSizes,
	}
}

func toBaseSizeResponse(size *model.Size) *productpb.BaseSizeResponse {
	return &productpb.BaseSizeResponse{
		Id:   size.ID,
		Name: size.Name,
	}
}

func toBaseVariantResponse(variant *model.Variant) *productpb.BaseVariantResponse {
	return &productpb.BaseVariantResponse{
		Id: variant.ID,
		Color: &productpb.BaseColorResponse{
			Id:   variant.ColorID,
			Name: variant.Color.Name,
		},
		Size: &productpb.BaseSizeResponse{
			Id:   variant.SizeID,
			Name: variant.Size.Name,
		},
		Inventory: &productpb.BaseInventoryResponse{
			Id:           variant.Inventory.ID,
			SoldQuantity: proto.Int64(int64(variant.Inventory.SoldQuantity)),
			Stock:        int64(variant.Inventory.Stock),
			IsStock:      &variant.Inventory.IsStock,
		},
	}
}

func toBaseImageResponse(image *model.Image) *productpb.BaseImageResponse {
	return &productpb.BaseImageResponse{
		Id: image.ID,
		Color: &productpb.BaseColorResponse{
			Id:   image.ColorID,
			Name: image.Color.Name,
		},
		Url:         image.Url,
		SortOrder:   int32(image.SortOrder),
		IsThumbnail: &image.IsThumbnail,
	}
}
