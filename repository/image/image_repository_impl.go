package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type imageRepositoryImpl struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepositoryImpl{db}
}

func (r *imageRepositoryImpl) Create(ctx context.Context, image *model.Image) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return err
	}

	return nil
}

func (r *imageRepositoryImpl) CreateAllTx(ctx context.Context, tx *gorm.DB, images []*model.Image) error {
	if err := tx.WithContext(ctx).Create(&images).Error; err != nil {
		return err
	}

	return nil
}

func (r *imageRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Image, error) {
	var images []*model.Image
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&images).Error; err != nil {
		return nil, err
	}

	return images, nil
}

func (r *imageRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Image{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrImageNotFound
	}

	return nil
}

func (r *imageRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Image{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *imageRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	return r.DeleteAllByIDTx(ctx, r.db, ids)
}

func (r *imageRepositoryImpl) DeleteAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) error {
	if err := tx.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Image{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *imageRepositoryImpl) UpdateFileID(ctx context.Context, id string, fileID string) error {
	result := r.db.WithContext(ctx).Model(&model.Image{}).Where("id = ?", id).Update("file_id", fileID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrImageNotFound
	}

	return nil
}
