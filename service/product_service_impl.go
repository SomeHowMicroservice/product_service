package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"github.com/SomeHowMicroservice/shm-be/product/mq"
	productpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/product"
	userpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/user"
	categoryRepo "github.com/SomeHowMicroservice/shm-be/product/repository/category"
	colorRepo "github.com/SomeHowMicroservice/shm-be/product/repository/color"
	imageRepo "github.com/SomeHowMicroservice/shm-be/product/repository/image"
	inventoryRepo "github.com/SomeHowMicroservice/shm-be/product/repository/inventory"
	productRepo "github.com/SomeHowMicroservice/shm-be/product/repository/product"
	sizeRepo "github.com/SomeHowMicroservice/shm-be/product/repository/size"
	tagRepo "github.com/SomeHowMicroservice/shm-be/product/repository/tag"
	variantRepo "github.com/SomeHowMicroservice/shm-be/product/repository/variant"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type productServiceImpl struct {
	cfg           *config.Config
	db            *gorm.DB
	userClient    userpb.UserServiceClient
	mqChannel     *amqp091.Channel
	categoryRepo  categoryRepo.CategoryRepository
	productRepo   productRepo.ProductRepository
	tagRepo       tagRepo.TagRepository
	colorRepo     colorRepo.ColorRepository
	sizeRepo      sizeRepo.SizeRepository
	variantRepo   variantRepo.VariantRepository
	inventoryRepo inventoryRepo.InventoryRepository
	imageRepo     imageRepo.ImageRepository
}

func NewProductService(cfg *config.Config, db *gorm.DB, userClient userpb.UserServiceClient, mqChannel *amqp091.Channel, categoryRepo categoryRepo.CategoryRepository, productRepo productRepo.ProductRepository, tagRepo tagRepo.TagRepository, colorRepo colorRepo.ColorRepository, sizeRepo sizeRepo.SizeRepository, variantRepo variantRepo.VariantRepository, inventoryRepo inventoryRepo.InventoryRepository, imageRepo imageRepo.ImageRepository) ProductService {
	return &productServiceImpl{
		cfg,
		db,
		userClient,
		mqChannel,
		categoryRepo,
		productRepo,
		tagRepo,
		colorRepo,
		sizeRepo,
		variantRepo,
		inventoryRepo,
		imageRepo,
	}
}

func (s *productServiceImpl) CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (string, error) {
	if req.Slug == nil {
		slug := common.GenerateSlug(req.Name)
		req.Slug = &slug
	}

	var parents []*model.Category
	var err error
	if len(req.ParentIds) > 0 {
		if err = s.validateParentRelations(ctx, req.ParentIds); err != nil {
			return "", err
		}

		parents, err = s.categoryRepo.FindAllByID(ctx, req.ParentIds)
		if err != nil {
			return "", fmt.Errorf("tìm kiếm danh mục sản phẩm cha thất bại: %w", err)
		}
		if len(parents) != len(req.ParentIds) {
			return "", common.ErrHasCategoryNotFound
		}
	}

	category := &model.Category{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        *req.Slug,
		Parents:     parents,
		CreatedByID: req.UserId,
		UpdatedByID: req.UserId,
	}
	if err = s.categoryRepo.Create(ctx, category); err != nil {
		if isUniqueViolation(err) {
			return "", common.ErrSlugAlreadyExists
		}
		return "", fmt.Errorf("tạo danh mục sản phẩm thất bại: %w", err)
	}

	return category.ID, nil
}

func (s *productServiceImpl) GetCategoryTree(ctx context.Context) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAllWithParentsAndChildren(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả danh mục sản phẩm thất bại: %w", err)
	}

	catMap := make(map[string]*model.Category)
	for _, c := range categories {
		c.Children = nil
		catMap[c.ID] = c
	}

	var roots []*model.Category
	for _, c := range categories {
		if len(c.Parents) == 0 {
			roots = append(roots, c)
		} else {
			for _, p := range c.Parents {
				if parent, ok := catMap[p.ID]; ok {
					parent.Children = append(parent.Children, c)
				}
			}
		}
	}

	return roots, nil
}

func (s *productServiceImpl) GetCategoriesNoProduct(ctx context.Context) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAllWithProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách danh mục sản phẩm thất bại: %w", err)
	}

	var noProductCategories []*model.Category
	for _, cat := range categories {
		if len(cat.Products) == 0 {
			noProductCategories = append(noProductCategories, cat)
		}
	}

	return noProductCategories, nil
}

func (s *productServiceImpl) GetProductBySlug(ctx context.Context, productSlug string) (*model.Product, error) {
	product, err := s.productRepo.FindBySlugWithDetails(ctx, productSlug)
	if err != nil {
		return nil, fmt.Errorf("lấy sản phẩm thất bại: %w", err)
	}
	if product == nil {
		return nil, common.ErrProductNotFound
	}

	return product, nil
}

func (s *productServiceImpl) CreateColor(ctx context.Context, req *productpb.CreateColorRequest) (string, error) {
	slug := common.GenerateSlug(req.Name)

	color := &model.Color{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        slug,
		CreatedByID: req.UserId,
		UpdatedByID: req.UserId,
	}
	if err := s.colorRepo.Create(ctx, color); err != nil {
		if isUniqueViolation(err) {
			return "", common.ErrColorAlreadyExists
		}
		return "", fmt.Errorf("tạo màu sắc thất bại: %w", err)
	}

	return color.ID, nil
}

func (s *productServiceImpl) CreateSize(ctx context.Context, req *productpb.CreateSizeRequest) (string, error) {
	slug := common.GenerateSlug(req.Name)

	size := &model.Size{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        slug,
		CreatedByID: req.UserId,
		UpdatedByID: req.UserId,
	}
	if err := s.sizeRepo.Create(ctx, size); err != nil {
		if isUniqueViolation(err) {
			return "", common.ErrSizeAlreadyExists
		}
		return "", fmt.Errorf("tạo size thất bại: %w", err)
	}

	return size.ID, nil
}

func (s *productServiceImpl) GetProductsByCategory(ctx context.Context, categorySlug string) ([]*model.Product, error) {
	exists, err := s.categoryRepo.ExistsBySlug(ctx, categorySlug)
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
	}
	if !exists {
		return nil, common.ErrCategoryNotFound
	}

	products, err := s.productRepo.FindAllByCategorySlug(ctx, categorySlug)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách sản phẩm theo danh mục thất bại: %w", err)
	}

	return products, nil
}

func (s *productServiceImpl) CreateTag(ctx context.Context, req *productpb.CreateTagRequest) (string, error) {
	slug := common.GenerateSlug(req.Name)

	tag := &model.Tag{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Slug:        slug,
		CreatedByID: req.UserId,
		UpdatedByID: req.UserId,
	}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		if isUniqueViolation(err) {
			return "", common.ErrTagAlreadyExists
		}
		return "", fmt.Errorf("tạo nhãn sản phẩm thất bại: %w", err)
	}

	return tag.ID, nil
}

func (s *productServiceImpl) GetAllCategories(ctx context.Context) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả danh mục sản phẩm thất bại: %w", err)
	}

	return categories, nil
}

func (s *productServiceImpl) GetCategoryByID(ctx context.Context, categoryID string) (*productpb.CategoryAdminDetailsResponse, error) {
	category, err := s.categoryRepo.FindByIDWithParentsAndProducts(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
	}
	if category == nil {
		return nil, common.ErrCategoryNotFound
	}

	cRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: category.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	uRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: category.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	productResponses := toBaseProductResponse(category)

	return toCategoryAdminDetailsResponse(category, productResponses, cRes, uRes), nil
}

func (s *productServiceImpl) UpdateCategory(ctx context.Context, req *productpb.UpdateCategoryRequest) (*productpb.CategoryAdminDetailsResponse, error) {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		category, err := s.categoryRepo.FindByIDWithParentsTx(ctx, tx, req.Id)
		if err != nil {
			if strings.Contains(err.Error(), "lock") {
				return fmt.Errorf("danh mục đang được cập nhật bởi người dùng khác, vui lòng thử lại: %w", err)
			}
			return fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
		}
		if category == nil {
			return common.ErrCategoryNotFound
		}

		updateData := map[string]interface{}{}
		if category.Name != req.Name {
			updateData["name"] = req.Name
		}
		if category.Slug != req.Slug {
			updateData["slug"] = req.Slug
		}
		if category.UpdatedByID != req.UserId {
			updateData["updated_by_id"] = req.UserId
		}

		if len(updateData) > 0 {
			if err = s.categoryRepo.UpdateTx(ctx, tx, category.ID, updateData); err != nil {
				if isUniqueViolation(err) {
					return common.ErrSlugAlreadyExists
				}
				return fmt.Errorf("cập nhật danh mục sản phẩm thất bại: %w", err)
			}
		}

		parentIDs := getIDsFromCategories(category.Parents)
		if !slices.Equal(parentIDs, req.ParentIds) {
			if err = s.validateParentRelations(ctx, req.ParentIds); err != nil {
				return err
			}
			if err = s.validateNoCycleWithCategory(ctx, req.Id, req.ParentIds); err != nil {
				return err
			}

			parents, err := s.categoryRepo.FindAllByIDTx(ctx, tx, req.ParentIds)
			if err != nil {
				return fmt.Errorf("tìm kiếm danh mục sản phẩm cha thất bại: %w", err)
			}

			if len(parents) != len(req.ParentIds) {
				return common.ErrHasCategoryNotFound
			}

			if err = s.categoryRepo.UpdateParentsTx(ctx, tx, category, parents); err != nil {
				return fmt.Errorf("cập nhật danh mục cha thất bại: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	category, err := s.categoryRepo.FindByIDWithParentsAndProducts(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
	}
	if category == nil {
		return nil, common.ErrCategoryNotFound
	}

	cRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: category.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	uRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: category.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	productResponses := toBaseProductResponse(category)

	return toCategoryAdminDetailsResponse(category, productResponses, cRes, uRes), nil
}

func (s *productServiceImpl) GetAllColorsAdmin(ctx context.Context) (*productpb.ColorsAdminResponse, error) {
	colors, err := s.colorRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách màu sắc sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, color := range colors {
		userIDMap[color.CreatedByID] = struct{}{}
		userIDMap[color.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var colorResponses []*productpb.ColorAdminResponse
	for _, color := range colors {
		colorResponses = append(colorResponses, &productpb.ColorAdminResponse{
			Id:        color.ID,
			Name:      color.Name,
			CreatedAt: color.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       color.CreatedByID,
				Username: userMap[color.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[color.CreatedByID].Profile.Id,
					FirstName: userMap[color.CreatedByID].Profile.FirstName,
					LastName:  userMap[color.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: color.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       color.UpdatedByID,
				Username: userMap[color.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[color.UpdatedByID].Profile.Id,
					FirstName: userMap[color.UpdatedByID].Profile.FirstName,
					LastName:  userMap[color.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.ColorsAdminResponse{
		Colors: colorResponses,
	}, nil
}

func (s *productServiceImpl) GetAllSizesAdmin(ctx context.Context) (*productpb.SizesAdminResponse, error) {
	sizes, err := s.sizeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách size sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, size := range sizes {
		userIDMap[size.CreatedByID] = struct{}{}
		userIDMap[size.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var sizeResponses []*productpb.SizeAdminResponse
	for _, size := range sizes {
		sizeResponses = append(sizeResponses, &productpb.SizeAdminResponse{
			Id:        size.ID,
			Name:      size.Name,
			CreatedAt: size.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       size.CreatedByID,
				Username: userMap[size.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[size.CreatedByID].Profile.Id,
					FirstName: userMap[size.CreatedByID].Profile.FirstName,
					LastName:  userMap[size.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: size.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       size.UpdatedByID,
				Username: userMap[size.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[size.UpdatedByID].Profile.Id,
					FirstName: userMap[size.UpdatedByID].Profile.FirstName,
					LastName:  userMap[size.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.SizesAdminResponse{
		Sizes: sizeResponses,
	}, nil
}

func (s *productServiceImpl) GetAllTagsAdmin(ctx context.Context) (*productpb.TagsAdminResponse, error) {
	tags, err := s.tagRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách tag sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, tag := range tags {
		userIDMap[tag.CreatedByID] = struct{}{}
		userIDMap[tag.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var tagResponses []*productpb.TagAdminResponse
	for _, tag := range tags {
		tagResponses = append(tagResponses, &productpb.TagAdminResponse{
			Id:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       tag.CreatedByID,
				Username: userMap[tag.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[tag.CreatedByID].Profile.Id,
					FirstName: userMap[tag.CreatedByID].Profile.FirstName,
					LastName:  userMap[tag.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       tag.UpdatedByID,
				Username: userMap[tag.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[tag.UpdatedByID].Profile.Id,
					FirstName: userMap[tag.UpdatedByID].Profile.FirstName,
					LastName:  userMap[tag.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.TagsAdminResponse{
		Tags: tagResponses,
	}, nil
}

func (s *productServiceImpl) UpdateTag(ctx context.Context, req *productpb.UpdateTagRequest) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		tag, err := s.tagRepo.FindByIDTx(ctx, tx, req.Id)
		if err != nil {
			return fmt.Errorf("tìm kiếm tag sản phẩm thất bại: %w", err)
		}
		if tag == nil {
			return common.ErrTagNotFound
		}

		updateData := map[string]interface{}{}
		if tag.Name != req.Name {
			updateData["name"] = req.Name
			updateData["slug"] = common.GenerateSlug(req.Name)
		}
		if tag.UpdatedByID != req.UserId {
			updateData["updated_by_id"] = req.UserId
		}

		if len(updateData) > 0 {
			if err = s.tagRepo.UpdateTx(ctx, tx, tag.ID, updateData); err != nil {
				if isUniqueViolation(err) {
					return common.ErrTagAlreadyExists
				}
				return fmt.Errorf("cập nhật tag sản phẩm thất bại: %w", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *productServiceImpl) UpdateColor(ctx context.Context, req *productpb.UpdateColorRequest) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		color, err := s.colorRepo.FindByIDTx(ctx, tx, req.Id)
		if err != nil {
			if strings.Contains(err.Error(), "lock") {
				return fmt.Errorf("màu sắc đang được cập nhật bởi người dùng khác, vui lòng thử lại: %w", err)
			}
			return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
		}
		if color == nil {
			return common.ErrColorNotFound
		}

		updateData := map[string]interface{}{}
		if color.Name != req.Name {
			updateData["name"] = req.Name
			updateData["slug"] = common.GenerateSlug(req.Name)
		}
		if color.UpdatedByID != req.UserId {
			updateData["updated_by_id"] = req.UserId
		}

		if len(updateData) > 0 {
			if err = s.colorRepo.UpdateTx(ctx, tx, color.ID, updateData); err != nil {
				if isUniqueViolation(err) {
					return common.ErrColorAlreadyExists
				}
				return fmt.Errorf("cập nhật màu sắc sản phẩm thất bại: %w", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *productServiceImpl) UpdateSize(ctx context.Context, req *productpb.UpdateSizeRequest) error {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		size, err := s.sizeRepo.FindByIDTx(ctx, tx, req.Id)
		if err != nil {
			if strings.Contains(err.Error(), "lock") {
				return fmt.Errorf("size đang được cập nhật bởi người dùng khác, vui lòng thử lại: %w", err)
			}
			return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
		}
		if size == nil {
			return common.ErrSizeNotFound
		}

		updateData := map[string]interface{}{}
		if size.Name != req.Name {
			updateData["name"] = req.Name
			updateData["slug"] = common.GenerateSlug(req.Name)
		}
		if size.UpdatedByID != req.UserId {
			updateData["updated_by_id"] = req.UserId
		}

		if len(updateData) > 0 {
			if err = s.sizeRepo.UpdateTx(ctx, tx, size.ID, updateData); err != nil {
				if isUniqueViolation(err) {
					return common.ErrSizeAlreadyExists
				}
				return fmt.Errorf("cập nhật kích cỡ sản phẩm thất bại: %w", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *productServiceImpl) GetAllColors(ctx context.Context) ([]*model.Color, error) {
	colors, err := s.colorRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả màu sắc sản phẩm thất bại: %w", err)
	}

	return colors, nil
}

func (s *productServiceImpl) GetAllSizes(ctx context.Context) ([]*model.Size, error) {
	sizes, err := s.sizeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả size sản phẩm thất bại: %w", err)
	}

	return sizes, nil
}

func (s *productServiceImpl) GetAllTags(ctx context.Context) ([]*model.Tag, error) {
	tags, err := s.tagRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả tag sản phẩm thất bại: %w", err)
	}

	return tags, nil
}

func (s *productServiceImpl) GetCategoriesNoChild(ctx context.Context) ([]*model.Category, error) {
	categories, err := s.categoryRepo.FindAllWithChildren(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy tất cả danh mục sản phẩm thất bại: %w", err)
	}

	var noChildCategories []*model.Category
	for _, c := range categories {
		if len(c.Children) == 0 {
			noChildCategories = append(noChildCategories, c)
		}
	}

	return noChildCategories, nil
}

func (s *productServiceImpl) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (string, error) {
	slug := common.GenerateSlug(req.Title)

	var categories []*model.Category
	var err error
	if len(req.CategoryIds) > 0 {
		categories, err = s.categoryRepo.FindAllByIDWithChildren(ctx, req.CategoryIds)
		if err != nil {
			return "", fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
		}
		if len(categories) != len(req.CategoryIds) {
			return "", common.ErrHasCategoryNotFound
		}

		for _, c := range categories {
			if len(c.Children) > 0 {
				return "", fmt.Errorf("danh mục %s có danh mục con, không thể được gán cho sản phẩm", c.Name)
			}
		}
	}

	var tags []*model.Tag
	if len(req.TagIds) > 0 {
		tags, err = s.tagRepo.FindAllByID(ctx, req.TagIds)
		if err != nil {
			return "", fmt.Errorf("tìm kiếm tag sản phẩm thất bại: %w", err)
		}
		if len(tags) != len(req.TagIds) {
			return "", common.ErrHasTagNotFound
		}
	}

	var startSale, endSale *time.Time
	if req.StartSale != nil && req.EndSale != nil {
		parsedStartSale, err := common.ParseDate(*req.StartSale)
		if err != nil {
			return "", fmt.Errorf("chuyển đổi kiểu dữ liệu thời gian bắt đầu giảm giá thất bại: %w", err)
		}
		startSale = &parsedStartSale

		parsedEndSale, err := common.ParseDate(*req.EndSale)
		if err != nil {
			return "", fmt.Errorf("chuyển đổi kiểu dữ liệu thời gian kết thúc giảm giá thất bại: %w", err)
		}
		endSale = &parsedEndSale
	}

	product := &model.Product{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Slug:        slug,
		Description: req.Description,
		Price:       req.Price,
		IsActive:    req.IsActive,
		IsSale:      req.IsSale,
		SalePrice:   req.SalePrice,
		StartSale:   startSale,
		EndSale:     endSale,
		Categories:  categories,
		Tags:        tags,
		CreatedByID: req.UserId,
		UpdatedByID: req.UserId,
	}

	variants := make([]*model.Variant, 0, len(req.Variants))
	for _, v := range req.Variants {
		exists, err := s.variantRepo.ExistsBySKU(ctx, v.Sku)
		if err != nil {
			return "", fmt.Errorf("kiểm tra mã SKU biến thể thất bại: %w", err)
		}
		if exists {
			return "", common.ErrSKUAlreadyExists
		}

		variant := &model.Variant{
			ID:        uuid.NewString(),
			ProductID: product.ID,
			SKU:       v.Sku,
			ColorID:   v.ColorId,
			SizeID:    v.SizeId,
			Inventory: &model.Inventory{
				ID:       uuid.NewString(),
				Quantity: int(v.Quantity),
			},
		}
		variant.Inventory.SetStock()

		variants = append(variants, variant)
	}
	product.Variants = variants

	images := make([]*model.Image, 0, len(req.Images))
	for _, img := range req.Images {
		ext := strings.ToLower(filepath.Ext(img.FileName))
		if ext == "" {
			ext = ".jpg"
		}
		fileName := fmt.Sprintf("%s-%s_%d%s", product.Slug, img.ColorId, img.SortOrder, ext)

		imageUrl := fmt.Sprintf("%s/%s/%s", s.cfg.ImageKit.URLEndpoint, s.cfg.ImageKit.Folder, fileName)
		image := &model.Image{
			ID:          uuid.NewString(),
			ProductID:   product.ID,
			ColorID:     img.ColorId,
			Url:         imageUrl,
			IsThumbnail: img.IsThumbnail,
			SortOrder:   int(img.SortOrder),
		}

		uploadFileRequest := &common.Base64UploadRequest{
			ImageID:    image.ID,
			Base64Data: img.Base64Data,
			FileName:   fileName,
			Folder:     s.cfg.ImageKit.Folder,
		}

		body, err := json.Marshal(uploadFileRequest)
		if err != nil {
			return "", fmt.Errorf("marshal json thất bại: %w", err)
		}

		if err = mq.PublishMessage(s.mqChannel, common.Exchange, common.UploadRoutingKey, body); err != nil {
			return "", fmt.Errorf("đẩy tin nhắn upload ảnh thất bại: %w", err)
		}

		images = append(images, image)
	}
	product.Images = images

	if err = s.productRepo.Create(ctx, product); err != nil {
		if isUniqueViolation(err) {
			return "", common.ErrSlugAlreadyExists
		}
		return "", fmt.Errorf("tạo sản phẩm thất bại: %w", err)
	}

	return product.ID, nil
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, productID string) (*productpb.ProductAdminDetailsResponse, error) {
	product, err := s.productRepo.FindByIDWithDetails(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if product == nil {
		return nil, common.ErrProductNotFound
	}

	cRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	uRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	return toProductAdminDetailsResponse(product, cRes, uRes), nil
}

func (s *productServiceImpl) GetAllProductsAdmin(ctx context.Context, req *productpb.GetAllProductsAdminRequest) ([]*model.Product, *common.PaginationMeta, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit > 100 {
		req.Limit = 100
	}

	query := &common.PaginationQuery{
		Page:       int(req.Page),
		Limit:      int(req.Limit),
		Sort:       req.Sort,
		Search:     req.Search,
		Order:      req.Order,
		IsActive:   req.IsActive,
		CategoryID: req.CategoryId,
	}

	products, total, err := s.productRepo.FindAllPaginatedWithCategoriesAndThumbnail(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("lấy tất cả sản phẩm thất bại: %w", err)
	}

	totalPages := int(total) / query.Limit
	if int(total)%query.Limit != 0 {
		totalPages++
	}

	meta := &common.PaginationMeta{
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return products, meta, nil
}

func (s *productServiceImpl) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductAdminDetailsResponse, error) {
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		product, err := s.productRepo.FindByIDWithCategoriesAndTagsTx(ctx, tx, req.Id)
		if err != nil {
			if strings.Contains(err.Error(), "lock") {
				return fmt.Errorf("sản phẩm đang được cập nhật bởi người dùng khác, vui lòng thử lại: %w", err)
			}
			return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
		}
		if product == nil {
			return common.ErrProductNotFound
		}

		updateProductData := map[string]interface{}{}
		if req.Title != nil && *req.Title != product.Title {
			updateProductData["title"] = req.Title
			updateProductData["slug"] = common.GenerateSlug(*req.Title)
		}
		if req.Description != nil && *req.Description != product.Description {
			updateProductData["description"] = req.Description
		}
		if req.Price != nil && *req.Price != product.Price {
			updateProductData["price"] = req.Price
		}
		if req.IsActive != nil && *req.IsActive != product.IsActive {
			updateProductData["is_active"] = req.IsActive
		}
		if req.IsSale != nil && *req.IsSale != product.IsSale {
			if !*req.IsSale {
				updateProductData["sale_price"] = nil
				updateProductData["start_sale"] = nil
				updateProductData["end_sale"] = nil
			}
			updateProductData["is_sale"] = req.IsSale
		}
		if req.SalePrice != nil && product.SalePrice != req.SalePrice {
			updateProductData["sale_price"] = req.SalePrice
		}
		if req.StartSale != nil {
			parsedStartSale, err := common.ParseDate(*req.StartSale)
			if err != nil {
				return fmt.Errorf("chuyển đổi kiểu dữ liệu thời gian bắt đầu giảm giá thất bại: %w", err)
			}
			if product.StartSale == nil || !parsedStartSale.Equal(*product.StartSale) {
				updateProductData["start_sale"] = parsedStartSale
			}
		}
		if req.EndSale != nil {
			parsedEndSale, err := common.ParseDate(*req.EndSale)
			if err != nil {
				return fmt.Errorf("chuyển đổi kiểu dữ liệu thời gian kết thúc giảm giá thất bại: %w", err)
			}
			if product.EndSale == nil || !parsedEndSale.Equal(*product.EndSale) {
				updateProductData["end_sale"] = parsedEndSale
			}
		}
		if req.UserId != product.UpdatedByID {
			updateProductData["updated_by_id"] = req.UserId
		}

		if len(updateProductData) > 0 {
			if err = s.productRepo.UpdateTx(ctx, tx, product.ID, updateProductData); err != nil {
				if isUniqueViolation(err) {
					return common.ErrSlugAlreadyExists
				}
				return fmt.Errorf("cập nhật sản phẩm thất bại: %w", err)
			}
		}

		if len(req.CategoryIds) > 0 {
			categoryIDs := getIDsFromCategories(product.Categories)
			if !slices.Equal(req.CategoryIds, categoryIDs) {
				categories, err := s.categoryRepo.FindAllByIDWithChildren(ctx, req.CategoryIds)
				if err != nil {
					return fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
				}
				if len(categories) != len(req.CategoryIds) {
					return common.ErrHasCategoryNotFound
				}

				for _, c := range categories {
					if len(c.Children) > 0 {
						return fmt.Errorf("danh mục %s có danh mục con, không thể được gán cho sản phẩm", c.Name)
					}
				}

				if err = s.productRepo.UpdateCategoriesTx(ctx, tx, product, categories); err != nil {
					return fmt.Errorf("cập nhật danh sách danh mục sản phẩm thất bại: %w", err)
				}
			}
		}

		if len(req.TagIds) > 0 {
			tagIDs := getIDsFromTags(product.Tags)
			if !slices.Equal(tagIDs, req.TagIds) {
				tags, err := s.tagRepo.FindAllByID(ctx, req.TagIds)
				if err != nil {
					return fmt.Errorf("tìm kiếm tag sản phẩm thất bại: %w", err)
				}

				if len(tags) != len(req.TagIds) {
					return common.ErrHasTagNotFound
				}

				if err = s.productRepo.UpdateTagsTx(ctx, tx, product, tags); err != nil {
					return fmt.Errorf("cập nhật tag sản phẩm thất bại: %w", err)
				}
			}
		}

		if len(req.DeleteImageIds) > 0 {
			images, err := s.imageRepo.FindAllByID(ctx, req.DeleteImageIds)
			if err != nil {
				return fmt.Errorf("tìm kiếm danh sách hình ảnh thất bại: %w", err)
			}
			if len(images) != len(req.DeleteImageIds) {
				return common.ErrHasImageNotFound
			}

			if err = s.imageRepo.DeleteAllByIDTx(ctx, tx, req.DeleteImageIds); err != nil {
				return fmt.Errorf("xóa danh sách hình ảnh thất bại: %w", err)
			}

			for _, image := range images {
				body := []byte(image.FileID)
				if err := mq.PublishMessage(s.mqChannel, common.Exchange, common.DeleteRoutingKey, body); err != nil {
					return fmt.Errorf("đẩy tin nhắn xóa hình ảnh thất bại: %w", err)
				}
			}
		}

		if len(req.UpdateImages) > 0 {
			for _, image := range req.UpdateImages {
				updateData := map[string]interface{}{}
				if image.IsThumbnail != nil {
					updateData["is_thumbnail"] = image.IsThumbnail
				}
				if image.SortOrder != nil {
					updateData["sort_order"] = image.SortOrder
				}

				if len(updateData) > 0 {
					if err = s.imageRepo.UpdateTx(ctx, tx, image.Id, updateData); err != nil {
						return fmt.Errorf("cập nhật ảnh %s thất bại: %w", image.Id, err)
					}
				}
			}
		}

		if len(req.NewImages) > 0 {
			newImages := make([]*model.Image, 0, len(req.NewImages))

			for _, img := range req.NewImages {
				ext := strings.ToLower(filepath.Ext(img.FileName))
				if ext == "" {
					ext = ".jpg"
				}
				fileName := fmt.Sprintf("%s-%s_%d%s", product.Slug, img.ColorId, img.SortOrder, ext)

				imageUrl := fmt.Sprintf("%s/%s/%s", s.cfg.ImageKit.URLEndpoint, s.cfg.ImageKit.Folder, fileName)
				image := &model.Image{
					ID:          uuid.NewString(),
					ProductID:   product.ID,
					ColorID:     img.ColorId,
					Url:         imageUrl,
					IsThumbnail: img.IsThumbnail,
					SortOrder:   int(img.SortOrder),
				}

				uploadFileRequest := &common.Base64UploadRequest{
					ImageID:    image.ID,
					Base64Data: img.Base64Data,
					FileName:   fileName,
					Folder:     s.cfg.ImageKit.Folder,
				}

				body, err := json.Marshal(uploadFileRequest)
				if err != nil {
					return fmt.Errorf("marshal json thất bại: %w", err)
				}

				if err = mq.PublishMessage(s.mqChannel, common.Exchange, common.UploadRoutingKey, body); err != nil {
					return fmt.Errorf("đẩy tin nhắn upload ảnh thất bại: %w", err)
				}

				newImages = append(newImages, image)
			}

			if err = s.imageRepo.CreateAllTx(ctx, tx, newImages); err != nil {
				return fmt.Errorf("tạo ảnh sản phẩm thất bại: %w", err)
			}
		}

		if len(req.DeleteVariantIds) > 0 {
			variants, err := s.variantRepo.FindAllByID(ctx, req.DeleteVariantIds)
			if err != nil {
				return fmt.Errorf("lấy danh sách danh mục sản phẩm thất bại: %w", err)
			}
			if len(variants) != len(req.DeleteVariantIds) {
				return common.ErrHasVariantNotFound
			}

			if err = s.variantRepo.DeleteAllByIDTx(ctx, tx, req.DeleteVariantIds); err != nil {
				return fmt.Errorf("xóa các biến thể sản phẩm thất bại: %w", err)
			}
		}

		if len(req.UpdateVariants) > 0 {
			for _, variant := range req.UpdateVariants {
				updateData := map[string]interface{}{}
				if variant.Sku != nil {
					updateData["sku"] = variant.Sku
				}
				if variant.ColorId != nil {
					updateData["color_id"] = variant.ColorId
				}
				if variant.SizeId != nil {
					updateData["size_id"] = variant.SizeId
				}

				if len(updateData) > 0 {
					if err = s.variantRepo.UpdateTx(ctx, tx, variant.Id, updateData); err != nil {
						if isUniqueViolation(err) {
							return common.ErrSKUAlreadyExists
						}
						return fmt.Errorf("cập nhật biến thể sản phẩm %s thất bại: %w", variant.Id, err)
					}
				}

				if variant.Quantity != nil {
					updateData := map[string]interface{}{
						"quantity": int(*variant.Quantity),
						"stock":    gorm.Expr("quantity - sold_quantity"),
						"is_stock": gorm.Expr("CASE WHEN (quantity - sold_quantity) <= 5 THEN false ELSE true END"),
					}
					if err = s.inventoryRepo.UpdateByVariantIDTx(ctx, tx, variant.Id, updateData); err != nil {
						if errors.Is(err, common.ErrInventoryNotFound) {
							return err
						}
						return fmt.Errorf("cập nhật số lượng biến thể %s thất bại: %w", variant.Id, err)
					}
				}
			}
		}

		if len(req.NewVariants) > 0 {
			newVariants := make([]*model.Variant, 0, len(req.NewVariants))

			for _, v := range req.NewVariants {
				variant := &model.Variant{
					ID:        uuid.NewString(),
					ProductID: product.ID,
					SKU:       v.Sku,
					ColorID:   v.ColorId,
					SizeID:    v.SizeId,
					Inventory: &model.Inventory{
						ID:       uuid.NewString(),
						Quantity: int(v.Quantity),
					},
				}
				variant.Inventory.SetStock()
				newVariants = append(newVariants, variant)
			}

			if err := s.variantRepo.CreateAllTx(ctx, tx, newVariants); err != nil {
				if isUniqueViolation(err) {
					return common.ErrHasSKUAlreadyExists
				}
				return fmt.Errorf("tạo variant mới thất bại: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	product, err := s.productRepo.FindByIDWithDetails(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("lấy thông tin sản phẩm sau cập nhật thất bại: %w", err)
	}
	if product == nil {
		return nil, common.ErrProductNotFound
	}

	cRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	uRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	return toProductAdminDetailsResponse(product, cRes, uRes), nil
}

func (s *productServiceImpl) DeleteProduct(ctx context.Context, req *productpb.DeleteOneRequest) error {
	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err := s.productRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrProductNotFound) {
			return err
		}
		return fmt.Errorf("chuyển sản phẩm vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteProducts(ctx context.Context, req *productpb.DeleteManyRequest) error {
	products, err := s.productRepo.FindAllByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if len(products) != len(req.Ids) {
		return common.ErrHasProductNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err = s.productRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("chuyển danh sách sản phẩm vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteCategory(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error {
	if err := s.categoryRepo.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, common.ErrCategoryNotFound) {
			return err
		}
		return fmt.Errorf("xóa danh mục sản phẩm thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteCategories(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error {
	categories, err := s.categoryRepo.FindAllByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm danh mục sản phẩm thất bại: %w", err)
	}
	if len(categories) != len(req.Ids) {
		return common.ErrHasCategoryNotFound
	}

	if err = s.categoryRepo.DeleteAllByID(ctx, req.Ids); err != nil {
		return fmt.Errorf("xóa danh sách danh mục sản phẩm thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteColor(ctx context.Context, req *productpb.DeleteOneRequest) error {
	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err := s.colorRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrColorNotFound) {
			return err
		}
		return fmt.Errorf("chuyển màu sắc vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteSize(ctx context.Context, req *productpb.DeleteOneRequest) error {
	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err := s.sizeRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrSizeNotFound) {
			return err
		}
		return fmt.Errorf("chuyển kích cỡ vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteColors(ctx context.Context, req *productpb.DeleteManyRequest) error {
	colors, err := s.colorRepo.FindAllByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
	}
	if len(colors) != len(req.Ids) {
		return common.ErrHasColorNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err = s.colorRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("chuyển danh sách màu sắc vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteSizes(ctx context.Context, req *productpb.DeleteManyRequest) error {
	sizes, err := s.sizeRepo.FindAllByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
	}
	if len(sizes) != len(req.Ids) {
		return common.ErrHasSizeNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err = s.sizeRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("chuyển danh sách kích cỡ vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) GetDeletedProducts(ctx context.Context, req *productpb.GetAllProductsAdminRequest) ([]*model.Product, *common.PaginationMeta, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit > 100 {
		req.Limit = 100
	}

	query := &common.PaginationQuery{
		Page:       int(req.Page),
		Limit:      int(req.Limit),
		Sort:       req.Sort,
		Search:     req.Search,
		Order:      req.Order,
		IsActive:   req.IsActive,
		CategoryID: req.CategoryId,
	}

	products, total, err := s.productRepo.FindAllDeletedPaginatedWithCategoriesAndThumbnail(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("lấy tất cả sản phẩm đã xóa thất bại: %w", err)
	}

	totalPages := int(total) / query.Limit
	if int(total)%query.Limit != 0 {
		totalPages++
	}

	meta := &common.PaginationMeta{
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	return products, meta, nil
}

func (s *productServiceImpl) GetDeletedProductByID(ctx context.Context, productID string) (*productpb.ProductAdminDetailsResponse, error) {
	product, err := s.productRepo.FindDeletedByIDWithDetails(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if product == nil {
		return nil, common.ErrProductNotFound
	}

	cRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	uRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{
		Id: product.CreatedByID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	return toProductAdminDetailsResponse(product, cRes, uRes), nil
}

func (s *productServiceImpl) GetDeletedColors(ctx context.Context) (*productpb.ColorsAdminResponse, error) {
	colors, err := s.colorRepo.FindAllDeleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách màu sắc sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, color := range colors {
		userIDMap[color.CreatedByID] = struct{}{}
		userIDMap[color.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var colorResponses []*productpb.ColorAdminResponse
	for _, color := range colors {
		colorResponses = append(colorResponses, &productpb.ColorAdminResponse{
			Id:        color.ID,
			Name:      color.Name,
			CreatedAt: color.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       color.CreatedByID,
				Username: userMap[color.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[color.CreatedByID].Profile.Id,
					FirstName: userMap[color.CreatedByID].Profile.FirstName,
					LastName:  userMap[color.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: color.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       color.UpdatedByID,
				Username: userMap[color.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[color.UpdatedByID].Profile.Id,
					FirstName: userMap[color.UpdatedByID].Profile.FirstName,
					LastName:  userMap[color.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.ColorsAdminResponse{
		Colors: colorResponses,
	}, nil
}

func (s *productServiceImpl) GetDeletedSizes(ctx context.Context) (*productpb.SizesAdminResponse, error) {
	sizes, err := s.sizeRepo.FindAllDeleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách size sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, size := range sizes {
		userIDMap[size.CreatedByID] = struct{}{}
		userIDMap[size.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var sizeResponses []*productpb.SizeAdminResponse
	for _, size := range sizes {
		sizeResponses = append(sizeResponses, &productpb.SizeAdminResponse{
			Id:        size.ID,
			Name:      size.Name,
			CreatedAt: size.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       size.CreatedByID,
				Username: userMap[size.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[size.CreatedByID].Profile.Id,
					FirstName: userMap[size.CreatedByID].Profile.FirstName,
					LastName:  userMap[size.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: size.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       size.UpdatedByID,
				Username: userMap[size.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[size.UpdatedByID].Profile.Id,
					FirstName: userMap[size.UpdatedByID].Profile.FirstName,
					LastName:  userMap[size.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.SizesAdminResponse{
		Sizes: sizeResponses,
	}, nil
}

func (s *productServiceImpl) GetDeletedTags(ctx context.Context) (*productpb.TagsAdminResponse, error) {
	tags, err := s.tagRepo.FindAllDeleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách tag sản phẩm thất bại: %w", err)
	}

	userIDMap := map[string]struct{}{}
	for _, tag := range tags {
		userIDMap[tag.CreatedByID] = struct{}{}
		userIDMap[tag.UpdatedByID] = struct{}{}
	}

	var userIDs []string
	for id := range userIDMap {
		userIDs = append(userIDs, id)
	}

	userRes, err := s.userClient.GetUsersById(ctx, &userpb.GetUsersByIdRequest{
		Ids: userIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrHasUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	userMap := make(map[string]*userpb.UserPublicResponse)
	for _, user := range userRes.Users {
		userMap[user.Id] = user
	}

	var tagResponses []*productpb.TagAdminResponse
	for _, tag := range tags {
		tagResponses = append(tagResponses, &productpb.TagAdminResponse{
			Id:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt.Format(time.RFC3339),
			CreatedBy: &productpb.BaseUserResponse{
				Id:       tag.CreatedByID,
				Username: userMap[tag.CreatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[tag.CreatedByID].Profile.Id,
					FirstName: userMap[tag.CreatedByID].Profile.FirstName,
					LastName:  userMap[tag.CreatedByID].Profile.LastName,
				},
			},
			UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
			UpdatedBy: &productpb.BaseUserResponse{
				Id:       tag.UpdatedByID,
				Username: userMap[tag.UpdatedByID].Username,
				Profile: &productpb.BaseProfileResponse{
					Id:        userMap[tag.UpdatedByID].Profile.Id,
					FirstName: userMap[tag.UpdatedByID].Profile.FirstName,
					LastName:  userMap[tag.UpdatedByID].Profile.LastName,
				},
			},
		})
	}

	return &productpb.TagsAdminResponse{
		Tags: tagResponses,
	}, nil
}

func (s *productServiceImpl) DeleteTag(ctx context.Context, req *productpb.DeleteOneRequest) error {
	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err := s.tagRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrTagNotFound) {
			return err
		}
		return fmt.Errorf("chuyển tag vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) DeleteTags(ctx context.Context, req *productpb.DeleteManyRequest) error {
	tags, err := s.tagRepo.FindAllByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm tag thất bại: %w", err)
	}
	if len(tags) != len(req.Ids) {
		return common.ErrHasSizeNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    true,
		"updated_by_id": req.UserId,
	}
	if err = s.tagRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("chuyển danh sách tag vào thùng rác thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreProduct(ctx context.Context, req *productpb.RestoreOneRequest) error {
	product, err := s.productRepo.FindDeletedByID(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if product == nil {
		return common.ErrProductNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.productRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrProductNotFound) {
			return err
		}
		return fmt.Errorf("khôi phục sản phẩm thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreProducts(ctx context.Context, req *productpb.RestoreManyRequest) error {
	products, err := s.productRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if len(products) != len(req.Ids) {
		return common.ErrHasProductNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.productRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("khôi phục danh sách sản phẩm thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreColor(ctx context.Context, req *productpb.RestoreOneRequest) error {
	color, err := s.colorRepo.FindDeletedByID(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
	}
	if color == nil {
		return common.ErrColorNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.colorRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrColorNotFound) {
			return err
		}
		return fmt.Errorf("khôi phục màu sắc thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreColors(ctx context.Context, req *productpb.RestoreManyRequest) error {
	colors, err := s.colorRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
	}
	if len(colors) != len(req.Ids) {
		return common.ErrHasColorNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.colorRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("khôi phục danh sách màu sắc thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreSize(ctx context.Context, req *productpb.RestoreOneRequest) error {
	size, err := s.sizeRepo.FindDeletedByID(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("tìm kiếm kích cỡ thất bại: %w", err)
	}
	if size == nil {
		return common.ErrSizeNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.sizeRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrSizeNotFound) {
			return err
		}
		return fmt.Errorf("khôi phục kích cỡ thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreSizes(ctx context.Context, req *productpb.RestoreManyRequest) error {
	sizes, err := s.sizeRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm kích cỡ thất bại: %w", err)
	}
	if len(sizes) != len(req.Ids) {
		return common.ErrHasSizeNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.sizeRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("khôi phục danh sách kích cỡ thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreTag(ctx context.Context, req *productpb.RestoreOneRequest) error {
	tag, err := s.tagRepo.FindDeletedByID(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("tìm kiếm kích cỡ thất bại: %w", err)
	}
	if tag == nil {
		return common.ErrTagNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.tagRepo.Update(ctx, req.Id, updateData); err != nil {
		if errors.Is(err, common.ErrSizeNotFound) {
			return err
		}
		return fmt.Errorf("khôi phục tag thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) RestoreTags(ctx context.Context, req *productpb.RestoreManyRequest) error {
	tags, err := s.tagRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm tag thất bại: %w", err)
	}
	if len(tags) != len(req.Ids) {
		return common.ErrHasTagNotFound
	}

	updateData := map[string]interface{}{
		"is_deleted":    false,
		"updated_by_id": req.UserId,
	}
	if err = s.tagRepo.UpdateAllByID(ctx, req.Ids, updateData); err != nil {
		return fmt.Errorf("khôi phục danh sách tag thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteProduct(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error {
	product, err := s.productRepo.FindDeletedByIDWithImages(ctx, req.Id)
	if err != nil {
		return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if product == nil {
		return common.ErrProductNotFound
	}

	if err = s.productRepo.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, common.ErrProductNotFound) {
			return err
		}
		return fmt.Errorf("xóa sản phẩm thất bại: %w", err)
	}

	for _, image := range product.Images {
		body := []byte(image.FileID)
		if err := mq.PublishMessage(s.mqChannel, common.Exchange, common.DeleteRoutingKey, body); err != nil {
			return fmt.Errorf("publish delete image msg thất bại: %w", err)
		}
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteProducts(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error {
	products, err := s.productRepo.FindAllDeletedByIDWithImages(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm sản phẩm thất bại: %w", err)
	}
	if len(products) != len(req.Ids) {
		return common.ErrHasProductNotFound
	}

	if err = s.productRepo.DeleteAllByID(ctx, req.Ids); err != nil {
		return fmt.Errorf("xóa danh sách sản phẩm thất bại: %w", err)
	}

	imageFileIDs := []string{}
	seen := make(map[string]bool)
	for _, product := range products {
		for _, image := range product.Images {
			if !seen[image.FileID] {
				seen[image.FileID] = true
				imageFileIDs = append(imageFileIDs, image.FileID)
			}
		}
	}

	for _, fileID := range imageFileIDs {
		body := []byte(fileID)
		if err := mq.PublishMessage(s.mqChannel, common.Exchange, common.DeleteRoutingKey, body); err != nil {
			return fmt.Errorf("publish delete image msg thất bại: %w", err)
		}
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteColor(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error {
	if err := s.colorRepo.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, common.ErrColorNotFound) {
			return err
		}
		return fmt.Errorf("xóa màu sắc thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteColors(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error {
	colors, err := s.colorRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm màu sắc thất bại: %w", err)
	}
	if len(colors) != len(req.Ids) {
		return common.ErrHasColorNotFound
	}

	if err = s.colorRepo.DeleteAllByID(ctx, req.Ids); err != nil {
		return fmt.Errorf("xóa danh sách màu sắc thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteSize(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error {
	if err := s.sizeRepo.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, common.ErrSizeNotFound) {
			return err
		}
		return fmt.Errorf("xóa kích cỡ thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteSizes(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error {
	sizes, err := s.sizeRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm kích cỡ thất bại: %w", err)
	}
	if len(sizes) != len(req.Ids) {
		return common.ErrHasSizeNotFound
	}

	if err = s.sizeRepo.DeleteAllByID(ctx, req.Ids); err != nil {
		return fmt.Errorf("xóa danh sách kích cỡ thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteTag(ctx context.Context, req *productpb.PermanentlyDeleteOneRequest) error {
	if err := s.tagRepo.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, common.ErrTagNotFound) {
			return err
		}
		return fmt.Errorf("xóa tag thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) PermanentlyDeleteTags(ctx context.Context, req *productpb.PermanentlyDeleteManyRequest) error {
	tags, err := s.tagRepo.FindAllDeletedByID(ctx, req.Ids)
	if err != nil {
		return fmt.Errorf("tìm kiếm tag thất bại: %w", err)
	}
	if len(tags) != len(req.Ids) {
		return common.ErrHasTagNotFound
	}

	if err = s.tagRepo.DeleteAllByID(ctx, req.Ids); err != nil {
		return fmt.Errorf("xóa danh sách tag thất bại: %w", err)
	}

	return nil
}

func (s *productServiceImpl) validateParentRelations(ctx context.Context, parentIDs []string) error {
	if len(parentIDs) <= 1 {
		return nil
	}

	idSet := make(map[string]struct{}, len(parentIDs))
	for _, id := range parentIDs {
		idSet[id] = struct{}{}
	}

	for _, id := range parentIDs {
		ancestors, err := s.categoryRepo.GetAllAncestors(ctx, id)
		if err != nil {
			return fmt.Errorf("xác thực mối quan hệ thất bại: %w", err)
		}
		for _, anc := range ancestors {
			if _, ok := idSet[anc]; ok {
				return fmt.Errorf("ID %s và %s có mối quan hệ", anc, id)
			}
		}
	}

	return nil
}

func (s *productServiceImpl) validateNoCycleWithCategory(ctx context.Context, categoryID string, parentIDs []string) error {
	if len(parentIDs) == 0 {
		return nil
	}

	descendants, err := s.categoryRepo.GetAllDescendants(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("xác thực chu kỳ danh mục thất bại: %w", err)
	}
	descSet := make(map[string]struct{}, len(descendants))
	for _, d := range descendants {
		descSet[d] = struct{}{}
	}

	ancestors, err := s.categoryRepo.GetAllAncestors(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("xác thực chu kỳ danh mục thất bại: %w", err)
	}
	ancSet := make(map[string]struct{}, len(ancestors))
	for _, a := range ancestors {
		ancSet[a] = struct{}{}
	}

	for _, pid := range parentIDs {
		if _, ok := descSet[pid]; ok {
			return fmt.Errorf("ID cha %s không hợp lệ có thể tạo ra chu kỳ", pid)
		}
		if _, ok := ancSet[pid]; ok {
			return fmt.Errorf("ID cha %s không hợp lệ nó đã có mối quan hệ với danh mục được gán", pid)
		}
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

func getIDsFromTags(tags []*model.Tag) []string {
	var tagIDs []string
	for _, tag := range tags {
		tagIDs = append(tagIDs, tag.ID)
	}

	return tagIDs
}

func getIDsFromCategories(categories []*model.Category) []string {
	var categoryIDs []string
	for _, cat := range categories {
		categoryIDs = append(categoryIDs, cat.ID)
	}

	return categoryIDs
}

func toProductAdminDetailsResponse(product *model.Product, cRes *userpb.UserResponse, uRes *userpb.UserResponse) *productpb.ProductAdminDetailsResponse {
	var startSalePtr, endSalePtr *string
	if product.StartSale != nil {
		formatted := product.StartSale.Format("2006-01-02")
		startSalePtr = &formatted
	}
	if product.EndSale != nil {
		formatted := product.EndSale.Format("2006-01-02")
		endSalePtr = &formatted
	}

	return &productpb.ProductAdminDetailsResponse{
		Id:          product.ID,
		Title:       product.Title,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		IsActive:    &product.IsActive,
		IsSale:      &product.IsSale,
		SalePrice:   product.SalePrice,
		StartSale:   startSalePtr,
		EndSale:     endSalePtr,
		Categories:  toBaseCategoriesResponse(product.Categories),
		Tags:        toBaseTagsResponse(product.Tags),
		Variants:    toBaseVariantsResponse(product.Variants),
		Images:      toBaseImagesResponse(product.Images),
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
		CreatedBy: &productpb.BaseUserResponse{
			Id:       cRes.Id,
			Username: cRes.Username,
			Profile: &productpb.BaseProfileResponse{
				Id:        cRes.Profile.Id,
				FirstName: cRes.Profile.FirstName,
				LastName:  cRes.Profile.LastName,
			},
		},
		UpdatedBy: &productpb.BaseUserResponse{
			Id:       uRes.Id,
			Username: uRes.Username,
			Profile: &productpb.BaseProfileResponse{
				Id:        uRes.Profile.Id,
				FirstName: uRes.Profile.FirstName,
				LastName:  uRes.Profile.LastName,
			},
		},
	}
}

func toBaseTagsResponse(tags []*model.Tag) []*productpb.BaseTagResponse {
	var tagResponses []*productpb.BaseTagResponse
	for _, t := range tags {
		tagResponses = append(tagResponses, &productpb.BaseTagResponse{
			Id:   t.ID,
			Name: t.Name,
		})
	}
	return tagResponses
}

func toBaseVariantsResponse(variants []*model.Variant) []*productpb.BaseVariantResponse {
	var variantResponses []*productpb.BaseVariantResponse
	for _, v := range variants {
		var color *productpb.BaseColorResponse
		if v.Color != nil {
			color = &productpb.BaseColorResponse{
				Id:   v.Color.ID,
				Name: v.Color.Name,
			}
		}

		var size *productpb.BaseSizeResponse
		if v.Size != nil {
			size = &productpb.BaseSizeResponse{
				Id:   v.Size.ID,
				Name: v.Size.Name,
			}
		}

		variantResponses = append(variantResponses, &productpb.BaseVariantResponse{
			Id:    v.ID,
			Sku:   v.SKU,
			Color: color,
			Size:  size,
			Inventory: &productpb.BaseInventoryResponse{
				Id:           v.Inventory.ID,
				Quantity:     int64(v.Inventory.Quantity),
				SoldQuantity: proto.Int64(int64(v.Inventory.SoldQuantity)),
				Stock:        int64(v.Inventory.Stock),
				IsStock:      &v.Inventory.IsStock,
			},
		})
	}

	return variantResponses
}

func toBaseImagesResponse(images []*model.Image) []*productpb.BaseImageResponse {
	var imageResponses []*productpb.BaseImageResponse
	for _, img := range images {
		imageResponses = append(imageResponses, &productpb.BaseImageResponse{
			Id:          img.ID,
			Url:         img.Url,
			IsThumbnail: &img.IsThumbnail,
			SortOrder:   int32(img.SortOrder),
		})
	}
	return imageResponses
}

func toCategoryAdminDetailsResponse(category *model.Category, productResponses []*productpb.BaseProductResponse, cRes *userpb.UserResponse, uRes *userpb.UserResponse) *productpb.CategoryAdminDetailsResponse {
	return &productpb.CategoryAdminDetailsResponse{
		Id:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		Parents:   toBaseCategoriesResponse(category.Parents),
		CreatedAt: category.CreatedAt.Format(time.RFC3339),
		UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
		CreatedBy: &productpb.BaseUserResponse{
			Id:       cRes.Id,
			Username: cRes.Username,
			Profile: &productpb.BaseProfileResponse{
				Id:        cRes.Profile.Id,
				FirstName: cRes.Profile.FirstName,
				LastName:  cRes.Profile.LastName,
			},
		},
		UpdatedBy: &productpb.BaseUserResponse{
			Id:       uRes.Id,
			Username: uRes.Username,
			Profile: &productpb.BaseProfileResponse{
				Id:        uRes.Profile.Id,
				FirstName: uRes.Profile.FirstName,
				LastName:  uRes.Profile.LastName,
			},
		},
		Products: productResponses,
	}
}

func toBaseProductResponse(category *model.Category) []*productpb.BaseProductResponse {
	var productResponses []*productpb.BaseProductResponse
	for _, p := range category.Products {
		var thumb *productpb.BaseImageResponse
		for _, img := range p.Images {
			if img.IsThumbnail {
				thumb = &productpb.BaseImageResponse{
					Id:          img.ID,
					Url:         img.Url,
					IsThumbnail: &img.IsThumbnail,
				}
				break
			}
		}

		productResponses = append(productResponses, &productpb.BaseProductResponse{
			Id:    p.ID,
			Title: p.Title,
			Slug:  p.Slug,
			Image: thumb,
		})
	}

	return productResponses
}

func toBaseCategoriesResponse(categories []*model.Category) []*productpb.BaseCategoryResponse {
	var baseCategories []*productpb.BaseCategoryResponse
	for _, category := range categories {
		baseCategories = append(baseCategories, &productpb.BaseCategoryResponse{
			Id:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		})
	}
	return baseCategories
}
