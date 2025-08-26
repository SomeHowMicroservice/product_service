package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type variantRepositoryImpl struct {
	db *gorm.DB
}

func NewVariantRepository(db *gorm.DB) VariantRepository {
	return &variantRepositoryImpl{db}
}

func (r *variantRepositoryImpl) Create(ctx context.Context, variant *model.Variant) error {
	if err := r.db.WithContext(ctx).Create(variant).Error; err != nil {
		return err
	}

	return nil
}

func (r *variantRepositoryImpl) CreateAllTx(ctx context.Context, tx *gorm.DB, variants []*model.Variant) error {
	if err := tx.WithContext(ctx).Create(&variants).Error; err != nil {
		return err
	}

	return nil
}

func (r *variantRepositoryImpl) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Variant{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *variantRepositoryImpl) FindAllByID(ctx context.Context, ids []string) ([]*model.Variant, error) {
	var variants []*model.Variant
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&variants).Error; err != nil {
		return nil, err
	}

	return variants, nil
}

func (r *variantRepositoryImpl) UpdateTx(ctx context.Context, tx *gorm.DB, id string, updateData map[string]interface{}) error {
	if err := tx.WithContext(ctx).Model(&model.Variant{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *variantRepositoryImpl) DeleteAllByIDTx(ctx context.Context, tx *gorm.DB, ids []string) error {
	if err := tx.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Variant{}).Error; err != nil {
		return err
	}

	return nil
}
