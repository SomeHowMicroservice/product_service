package model

type Variant struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	SKU         string    `gorm:"type:varchar(50);uniqueIndex:variants_sku_key;not null" json:"sku"`
	ProductID   string    `gorm:"type:char(36);not null" json:"-"`
	ColorID     string    `gorm:"type:char(36);not null" json:"-"`
	SizeID      string    `gorm:"type:char(36);not null" json:"-"`

	Product   *Product   `gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"product"`
	Color     *Color     `gorm:"foreignKey:ColorID;references:ID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE" json:"color"`
	Size      *Size      `gorm:"foreignKey:SizeID;references:ID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE" json:"size"`
	Inventory *Inventory `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"inventory"`
}
