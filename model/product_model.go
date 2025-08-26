package model

import "time"

type Product struct {
	ID          string     `gorm:"type:char(36);primaryKey" json:"id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string     `gorm:"type:varchar(255);uniqueIndex:products_slug_key;not null" json:"slug"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Price       float32    `gorm:"type:decimal(10,2);not null" json:"price"`
	IsActive    bool       `gorm:"type:boolean;not null;default:true" json:"is_active"`
	IsSale      bool       `gorm:"type:boolean;not null" json:"is_sale"`
	SalePrice   *float32   `gorm:"type:decimal(10,2)" json:"sale_price"`
	StartSale   *time.Time `gorm:"type:date" json:"start_sale"`
	EndSale     *time.Time `gorm:"type:date" json:"end_sale"`
	IsDeleted   bool       `gorm:"type:boolean;not null;default:false" json:"is_deleted"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedByID string     `gorm:"type:char(36);not null" json:"created_by_id"`
	UpdatedByID string     `gorm:"type:char(36);not null" json:"updated_by_id"`

	Categories []*Category `gorm:"many2many:product_categories;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"categories"`
	Tags       []*Tag      `gorm:"many2many:product_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"tags"`
	Variants   []*Variant  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE" json:"variants"`
	Images     []*Image    `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE" json:"images"`
}
