package repository

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/gorm"
)

type inventoryRepositoryImpl struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepositoryImpl{db}
}

func (r *inventoryRepositoryImpl) UpdateByVariantIDTx(ctx context.Context, tx *gorm.DB, variantID string, updateData map[string]interface{}) error {
	result := tx.WithContext(ctx).Model(&model.Inventory{}).Where("variant_id = ?", variantID).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common.ErrInventoryNotFound
	}

	return nil
}