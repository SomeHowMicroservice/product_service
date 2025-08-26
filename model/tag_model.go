package model

import (
	"time"
)

type Tag struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(50);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(50);uniqueIndex:tags_slug_key;not null" json:"slug"`
	IsDeleted   bool      `gorm:"type:boolean;default:false" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedByID string    `gorm:"type:char(36);not null" json:"created_by_id"`
	UpdatedByID string    `gorm:"type:char(36);not null" json:"updated_by_id"`

	Products []*Product `gorm:"many2many:product_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"products"`
}
