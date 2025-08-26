package repository

import (
	"context"
	"errors"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tagRepositoryImpl struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepositoryImpl{db}
}

func (r *tagRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Tag{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *tagRepositoryImpl) Create(ctx context.Context, tag *model.Tag) error {
	if err := r.db.WithContext(ctx).Create(tag).Error; err != nil {
		return err
	}

	return nil
}

func (r *tagRepositoryImpl) FindAll(ctx context.Context) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := r.db.WithContext(ctx).Where("is_deleted = false").Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Tag, error) {
	return findByIDBase(ctx, r.db, id, nil)
}

func (r *tagRepositoryImpl) FindByIDTx(ctx context.Context, tx *gorm.DB, id string) (*model.Tag, error) {
	return findByIDBase(ctx, tx, id, &common.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsNoWait})
}

func (r *tagRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrTagNotFound
	}

	return nil
}

func (r *tagRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *tagRepositoryImpl) UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&model.Tag{}).Where("id IN ?", ids).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *tagRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = false", ids).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepositoryImpl) FindAllDeleted(ctx context.Context) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := r.db.WithContext(ctx).Where("is_deleted = true").Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepositoryImpl) FindDeletedByID(ctx context.Context, id string) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = true", id).First(&tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &tag, nil
}

func (r *tagRepositoryImpl) FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = true", ids).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Tag{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *tagRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrTagNotFound
	}

	return nil
}

func findByIDBase(ctx context.Context, tx *gorm.DB, id string, locking *common.Locking) (*model.Tag, error) {
	var tag model.Tag
	query := tx.WithContext(ctx)

	if locking != nil {
		query = query.Clauses(clause.Locking{Strength: locking.Strength, Options: locking.Options})
	}

	if err := query.Where("id = ? AND is_deleted = false", id).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &tag, nil
}
