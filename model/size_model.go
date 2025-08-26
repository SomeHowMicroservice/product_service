package model

import "time"

type Size struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(20);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(20);uniqueIndex:sizes_slug_key;not null" json:"slug"`
	IsDeleted   bool      `gorm:"type:boolean;default:false" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedByID string    `gorm:"type:char(36);not null" json:"created_by_id"`
	UpdatedByID string    `gorm:"type:char(36);not null" json:"updated_by_id"`

	Variants []*Variant `gorm:"foreignKey:SizeID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE" json:"-"`
}
