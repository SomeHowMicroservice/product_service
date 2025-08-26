package repository

import (
	"context"
	"errors"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sizeRepositoryImpl struct {
	db *gorm.DB
}

func NewSizeRepository(db *gorm.DB) SizeRepository {
	return &sizeRepositoryImpl{db}
}

func (r *sizeRepositoryImpl) Create(ctx context.Context, size *model.Size) error {
	if err := r.db.WithContext(ctx).Create(size).Error; err != nil {
		return err
	}

	return nil
}

func (r *sizeRepositoryImpl) FindAll(ctx context.Context) ([]*model.Size, error) {
	var sizes []*model.Size
	if err := r.db.WithContext(ctx).Where("is_deleted = false").Find(&sizes).Error; err != nil {
		return nil, err
	}

	return sizes, nil
}

func (r *sizeRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Size, error) {
	return findByIDBase(ctx, r.db, id, nil)
}

func (r *sizeRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Size{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *sizeRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Size{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrSizeNotFound
	}

	return nil
}

func (r *sizeRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Size{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *sizeRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Size, error) {
	var size []*model.Size
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = false", ids).Find(&size).Error; err != nil {
		return nil, err
	}

	return size, nil
}

func (r *sizeRepositoryImpl) FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Size, error) {
	var size []*model.Size
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = true", ids).Find(&size).Error; err != nil {
		return nil, err
	}

	return size, nil
}

func (r *sizeRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Size{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *sizeRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Size{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrSizeNotFound
	}

	return nil
}

func (r *sizeRepositoryImpl) UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&model.Size{}).Where("id IN ?", ids).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *sizeRepositoryImpl) FindDeletedByID(ctx context.Context, id string) (*model.Size, error) {
	var size model.Size
	if err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = true", id).First(&size).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &size, nil
}

func (r *sizeRepositoryImpl) FindAllDeleted(ctx context.Context) ([]*model.Size, error) {
	var sizes []*model.Size
	if err := r.db.WithContext(ctx).Where("is_deleted = true").Find(&sizes).Error; err != nil {
		return nil, err
	}

	return sizes, nil
}

func (r *sizeRepositoryImpl) FindByIDTx(ctx context.Context, tx *gorm.DB, id string) (*model.Size, error) {
	return findByIDBase(ctx, tx, id, &common.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsNoWait})
}

func findByIDBase(ctx context.Context, tx *gorm.DB, id string, locking *common.Locking) (*model.Size, error) {
	var size model.Size
	query := tx.WithContext(ctx)

	if locking != nil {
		query = query.Clauses(clause.Locking{Strength: locking.Strength, Options: locking.Options})
	}

	if err := query.Where("id = ? AND is_deleted = false", id).First(&size).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &size, nil
}
