package repository

import (
	"context"

	"github.com/SomeHowMicroservice/product/common"
	"github.com/SomeHowMicroservice/product/model"
	"gorm.io/gorm"
)

type imageRepositoryImpl struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepositoryImpl{db}
}

func (r *imageRepositoryImpl) CreateAllTx(ctx context.Context, tx *gorm.DB, images []*model.Image) error {
	return tx.WithContext(ctx).Create(&images).Error
}

func (r *imageRepositoryImpl) FindAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) ([]*model.Image, error) {
	var images []*model.Image
	if err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&images).Error; err != nil {
		return nil, err
	}

	return images, nil
}

func (r *imageRepositoryImpl) Update(ctx context.Context, id string, updateData map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Image{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrImageNotFound
	}

	return nil
}

func (r *imageRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]any) error {
	return tx.WithContext(ctx).Model(&model.Image{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *imageRepositoryImpl) DeleteAllByID(ctx context.Context, ids []string) error {
	return r.DeleteAllByIDTx(ctx, r.db, ids)
}

func (r *imageRepositoryImpl) DeleteAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) error {
	return tx.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Image{}).Error
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

func (r *imageRepositoryImpl) FindByProductIDWithColor(ctx context.Context, productID string) ([]*model.Image, error) {
	var images []*model.Image
	if err := r.db.WithContext(ctx).Preload("Color").Where("product_id = ?", productID).Find(&images).Error; err != nil {
		return nil, err
	}

	return images, nil
}
