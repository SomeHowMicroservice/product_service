package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepositoryImpl{db}
}

func (r *productRepositoryImpl) Create(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *productRepositoryImpl) FindBySlugWithDetails(ctx context.Context, slug string) (*model.Product, error) {
	return r.findBySlugBase(ctx, slug, "Categories", "Tags", "Variants", "Variants.Color", "Variants.Size", "Variants.Inventory", "Images", "Images.Color")
}

func (r *productRepositoryImpl) FindByIDWithDetails(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, false, nil,
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Tags", Scope: notDeleted},
		&common.Preload{Relation: "Variants"},
		&common.Preload{Relation: "Variants.Color", Scope: notDeleted},
		&common.Preload{Relation: "Variants.Size", Scope: notDeleted},
		&common.Preload{Relation: "Variants.Inventory"},
		&common.Preload{Relation: "Images"},
		&common.Preload{Relation: "Images.Color", Scope: notDeleted})
}

func (r *productRepositoryImpl) FindDeletedByIDWithDetails(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, true, nil,
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Tags", Scope: notDeleted},
		&common.Preload{Relation: "Variants"},
		&common.Preload{Relation: "Variants.Color", Scope: notDeleted},
		&common.Preload{Relation: "Variants.Size", Scope: notDeleted},
		&common.Preload{Relation: "Variants.Inventory"},
		&common.Preload{Relation: "Images"},
		&common.Preload{Relation: "Images.Color", Scope: notDeleted})
}

func (r *productRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *productRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, false, nil)
}

func (r *productRepositoryImpl) FindAllByCategorySlug(ctx context.Context, categorySlug string) ([]*model.Product, error) {
	var products []*model.Product
	if err := r.db.WithContext(ctx).Where("id IN (?)", r.db.Table("product_categories pc")).Select("pc.product_id").Joins("JOIN categories c ON c.id = pc.category_id").Where("c.slug = ? AND products.is_deleted = false", categorySlug).Preload("Categories").Preload("Variants").Preload("Variants.Color").Preload("Variants.Size").Preload("Variants.Inventory").Preload("Images").Preload("Images.Color").Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepositoryImpl) FindDeletedByID(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, true, nil)
}

func (r *productRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Product{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrProductNotFound
	}

	return nil
}

func (r *productRepositoryImpl) FindDeletedByIDWithImages(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, true, nil, &common.Preload{Relation: "Images"})
}

func (r *productRepositoryImpl) FindAllDeletedByIDWithImages(ctx context.Context, ids []string) ([]*model.Product, error) {
	return r.findAllByIDBase(ctx, ids, true, "Images")
}

func (r *productRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Product{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Product, error) {
	return r.findAllByIDBase(ctx, ids, true)
}

func (r *productRepositoryImpl) UpdateCategoriesTx(ctx context.Context, tx *gorm.DB, product *model.Product, categories []*model.Category) error {
	if err := tx.WithContext(ctx).Model(product).Association("Categories").Replace(categories); err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) FindByIDWithCategoriesAndTags(ctx context.Context, id string) (*model.Product, error) {
	return findByIDBase(ctx, r.db, id, false, nil,
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Tags"})
}

func (r *productRepositoryImpl) UpdateTagsTx(ctx context.Context, tx *gorm.DB, product *model.Product, tags []*model.Tag) error {
	if err := tx.WithContext(ctx).Model(product).Association("Tags").Replace(tags); err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrProductNotFound
	}

	return nil
}

func (r *productRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Product, error) {
	return r.findAllByIDBase(ctx, ids, false)
}

func (r *productRepositoryImpl) UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("id IN ?", ids).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *productRepositoryImpl) FindAllDeletedPaginatedWithCategoriesAndThumbnail(ctx context.Context, query *common.PaginationQuery) ([]*model.Product, int64, error) {
	return r.findAllPaginatedBase(ctx, true, query,
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Images", Scope: getThumbnail})
}

func (r *productRepositoryImpl) FindAllPaginatedWithCategoriesAndThumbnail(ctx context.Context, query *common.PaginationQuery) ([]*model.Product, int64, error) {
	return r.findAllPaginatedBase(ctx, false, query,
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Images", Scope: getThumbnail})
}

func (r *productRepositoryImpl) FindByIDWithCategoriesAndTagsTx(ctx context.Context, tx *gorm.DB, id string) (*model.Product, error) {
	return findByIDBase(ctx, tx, id, false,
		&common.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsNoWait},
		&common.Preload{Relation: "Categories"},
		&common.Preload{Relation: "Tags"})
}

func (r *productRepositoryImpl) findAllPaginatedBase(ctx context.Context, isDeleted bool, pQuery *common.PaginationQuery, preloads ...*common.Preload) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Product{})
	for _, preload := range preloads {
		if preload.Scope != nil {
			query = query.Preload(preload.Relation, preload.Scope)
		} else {
			query = query.Preload(preload.Relation)
		}
	}

	db := query.Where("is_deleted = ?", isDeleted)
	db = r.applyFilters(db, pQuery)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = r.applySorting(db, pQuery)

	offset := (pQuery.Page - 1) * pQuery.Limit
	if err := db.Offset(offset).Limit(pQuery.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepositoryImpl) applyFilters(db *gorm.DB, query *common.PaginationQuery) *gorm.DB {
	if query.Search != "" {
		searchTerm := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(title) LIKE ?", searchTerm)
	}

	if query.CategoryID != "" {
		db = db.Where("id IN (?)", r.db.Table("product_categories pc").Select("pc.product_id").Joins("JOIN categories c ON c.id = pc.category_id").Where("c.id = ?", query.CategoryID))
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	return db
}

func (r *productRepositoryImpl) applySorting(db *gorm.DB, query *common.PaginationQuery) *gorm.DB {
	if query.Sort == "" {
		query.Sort = "created_at"
	}
	if query.Order == "" {
		query.Order = "desc"
	}

	allowedSorts := map[string]bool{
		"price":      true,
		"created_at": true,
		"updated_at": true,
		"stock":      true,
	}

	if allowedSorts[query.Sort] {
		db = db.Order(query.Sort + " " + strings.ToUpper(query.Order))
	} else {
		db = db.Order("created_at DESC")
	}

	return db
}

func (r *productRepositoryImpl) findAllByIDBase(ctx context.Context, ids []string, isDeleted bool, preloads ...string) ([]*model.Product, error) {
	var products []*model.Product
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.Where("id IN ? AND is_deleted = ?", ids, isDeleted).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepositoryImpl) findBySlugBase(ctx context.Context, slug string, preloads ...string) (*model.Product, error) {
	var product model.Product
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.Scopes(notDeleted).Where("slug = ?", slug).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func findByIDBase(ctx context.Context, tx *gorm.DB, id string, isDeleted bool, locking *common.Locking, preloads ...*common.Preload) (*model.Product, error) {
	var product model.Product
	query := tx.WithContext(ctx)

	if locking != nil {
		query = query.Clauses(clause.Locking{Strength: locking.Strength, Options: locking.Options})
	}

	for _, preload := range preloads {
		if preload.Scope != nil {
			query = query.Preload(preload.Relation, preload.Scope)
		} else {
			query = query.Preload(preload.Relation)
		}
	}

	if err := query.Where("id = ? AND is_deleted = ?", id, isDeleted).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func notDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = false")
}

func getThumbnail(db *gorm.DB) *gorm.DB {
	return db.Where("is_thumbnail = true")
}
