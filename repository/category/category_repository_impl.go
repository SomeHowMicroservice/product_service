package repository

import (
	"context"
	"errors"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type categoryRepositoryImpl struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepositoryImpl{db}
}

func (r *categoryRepositoryImpl) Create(ctx context.Context, category *model.Category) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Category, error) {
	return r.findAllByIDBase(ctx, r.db, ids)
}

func (r *categoryRepositoryImpl) FindAllByIDWithChildren(ctx context.Context, ids []string) ([]*model.Category, error) {
	return r.findAllByIDBase(ctx, r.db, ids, "Children")
}

func (r *categoryRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.ExistsBySlugTx(ctx, r.db, slug)
}

func (r *categoryRepositoryImpl) ExistsBySlugTx(ctx context.Context, tx *gorm.DB, slug string) (bool, error) {
	var count int64
	if err := tx.WithContext(ctx).Model(&model.Category{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *categoryRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *categoryRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Category, error) {
	return r.findByIDBase(ctx, r.db, id, nil)
}

func (r *categoryRepositoryImpl) FindAllWithParentsAndChildren(ctx context.Context) ([]*model.Category, error) {
	return r.findAllBase(ctx, "Parents", "Children")
}

func (r *categoryRepositoryImpl) FindAllWithProducts(ctx context.Context) ([]*model.Category, error) {
	return r.findAllBase(ctx, "Products")
}

func (r *categoryRepositoryImpl) FindAll(ctx context.Context) ([]*model.Category, error) {
	return r.findAllBase(ctx)
}

func (r *categoryRepositoryImpl) FindByIDWithParentsAndProducts(ctx context.Context, id string) (*model.Category, error) {
	return r.findByIDBase(ctx, r.db, id, nil,
		&common.Preload{Relation: "Parents"},
		&common.Preload{Relation: "Products"},
		&common.Preload{Relation: "Products.Images", Scope: getThumbnail})
}

func (r *categoryRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepositoryImpl) FindByIDWithParents(ctx context.Context, id string) (*model.Category, error) {
	return r.findByIDBase(ctx, r.db, id, nil, &common.Preload{Relation: "Parents"})
}

func (r *categoryRepositoryImpl) UpdateParentsTx(ctx context.Context, tx *gorm.DB, category *model.Category, parents []*model.Category) error {
	if err := tx.WithContext(ctx).Model(category).Association("Parents").Replace(parents); err != nil {
		return err
	}

	return nil
}

func (r *categoryRepositoryImpl) FindAllWithChildren(ctx context.Context) ([]*model.Category, error) {
	return r.findAllBase(ctx, "Children")
}

func (r *categoryRepositoryImpl) FindAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) ([]*model.Category, error) {
	return r.findAllByIDBase(ctx, tx, ids)
}

func (r *categoryRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Category{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *categoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrCategoryNotFound
	}

	return nil
}

func (r *categoryRepositoryImpl) FindByIDWithParentsTx(ctx context.Context, tx *gorm.DB, id string) (*model.Category, error) {
	return r.findByIDBase(ctx, tx, id,
		&common.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsNoWait},
		&common.Preload{Relation: "Parents"})
}

func (r *categoryRepositoryImpl) GetAllAncestors(ctx context.Context, id string) ([]string, error) {
	var ancestors []string
	query := `
	WITH RECURSIVE ancestors AS (
		SELECT parent_id
		FROM category_parents
		WHERE child_id = ?
		UNION
		SELECT cp.parent_id
		FROM category_parents cp
		INNER JOIN ancestors a ON cp.child_id = a.parent_id
	)
	SELECT DISTINCT parent_id FROM ancestors;
	`
	if err := r.db.WithContext(ctx).Raw(query, id).Scan(&ancestors).Error; err != nil {
		return nil, err
	}

	return ancestors, nil
}

func (r *categoryRepositoryImpl) GetAllDescendants(ctx context.Context, id string) ([]string, error) {
	var descendants []string
	query := `
	WITH RECURSIVE descendants AS (
		SELECT child_id 
		FROM category_parents 
		WHERE parent_id = ?
		UNION
		SELECT cp.child_id 
		FROM category_parents cp
		INNER JOIN descendants d ON cp.parent_id = d.child_id
	)
	SELECT child_id FROM descendants;
	`
	if err := r.db.WithContext(ctx).Raw(query, id).Scan(&descendants).Error; err != nil {
		return nil, err
	}

	return descendants, nil
}

func (r *categoryRepositoryImpl) findByIDBase(ctx context.Context, tx *gorm.DB, id string, looking *common.Locking, preloads ...*common.Preload) (*model.Category, error) {
	var category model.Category
	query := tx.WithContext(ctx)

	if looking != nil {
		query.Clauses(clause.Locking{Strength: looking.Strength, Options: looking.Options})
	}

	for _, preload := range preloads {
		if preload.Scope != nil {
			query = query.Preload(preload.Relation, preload.Scope)
		} else {
			query = query.Preload(preload.Relation)
		}
	}

	if err := query.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepositoryImpl) findAllBase(ctx context.Context, preloads ...string) ([]*model.Category, error) {
	var categories []*model.Category
	query := r.db.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepositoryImpl) findAllByIDBase(ctx context.Context, tx *gorm.DB, ids []string, preloads ...string) ([]*model.Category, error) {
	var categories []*model.Category
	query := tx.WithContext(ctx)

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	if err := query.Where("id IN ?", ids).Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func getThumbnail(db *gorm.DB) *gorm.DB {
	return db.Where("is_thumbnail = true")
}
