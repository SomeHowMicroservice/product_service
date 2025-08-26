package repository

import (
	"context"
	"errors"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type colorRepositoryImpl struct {
	db *gorm.DB
}

func NewColorRepository(db *gorm.DB) ColorRepository {
	return &colorRepositoryImpl{db}
}

func (r *colorRepositoryImpl) Create(ctx context.Context, color *model.Color) error {
	if err := r.db.WithContext(ctx).Create(color).Error; err != nil {
		return err
	}

	return nil
}

func (r *colorRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Color{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *colorRepositoryImpl) FindAll(ctx context.Context) ([]*model.Color, error) {
	var colors []*model.Color
	if err := r.db.WithContext(ctx).Where("is_deleted = false").Find(&colors).Error; err != nil {
		return nil, err
	}

	return colors, nil
}

func (r *colorRepositoryImpl) FindAllDeleted(ctx context.Context) ([]*model.Color, error) {
	var colors []*model.Color
	if err := r.db.WithContext(ctx).Where("is_deleted = true").Find(&colors).Error; err != nil {
		return nil, err
	}

	return colors, nil
}

func (r *colorRepositoryImpl) FindAllDeletedByID(ctx context.Context, ids []string) ([]*model.Color, error) {
	var colors []*model.Color
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = true", ids).Find(&colors).Error; err != nil {
		return nil, err
	}

	return colors, nil
}

func (r *colorRepositoryImpl) FindDeletedByID(ctx context.Context, id string) (*model.Color, error) {
	var color model.Color
	if err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = true", id).First(&color).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &color, nil
}

func (r *colorRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Color, error) {
	return findByIDBase(ctx, r.db, id, nil)
}

func (r *colorRepositoryImpl) FindByIDTx(ctx context.Context, tx *gorm.DB, id string) (*model.Color, error) {
	return findByIDBase(ctx, tx, id, &common.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsNoWait})
}

func (r *colorRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Color{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *colorRepositoryImpl) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Color{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrColorNotFound
	}

	return nil
}

func (r *colorRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Color{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrColorNotFound
	}

	return nil
}

func (r *colorRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Color{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *colorRepositoryImpl) UpdateAllByID(ctx context.Context, ids []string, updateData map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&model.Color{}).Where("id IN ?", ids).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *colorRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Color, error) {
	var color []*model.Color
	if err := r.db.WithContext(ctx).Where("id IN ? AND is_deleted = false", ids).Find(&color).Error; err != nil {
		return nil, err
	}

	return color, nil
}

func findByIDBase(ctx context.Context, tx *gorm.DB, id string, locking *common.Locking) (*model.Color, error) {
	var color model.Color
	query := tx.WithContext(ctx)

	if locking != nil {
		query = query.Clauses(clause.Locking{Strength: locking.Strength, Options: locking.Options})
	}

	if err := query.Where("id = ? AND is_deleted = false", id).First(&color).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &color, nil
}
