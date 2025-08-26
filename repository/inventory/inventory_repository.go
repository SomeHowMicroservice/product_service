package repository

import (
	"context"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	UpdateByVariantIDTx(ctx context.Context, tx *gorm.DB, variantID string, updateData map[string]interface{}) error
}