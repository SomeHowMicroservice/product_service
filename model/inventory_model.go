package model

type Inventory struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	VariantID    string    `gorm:"type:char(36);uniqueIndex:inventories_variant_id_key;not null" json:"-"`
	Quantity     int       `gorm:"type:int" json:"quantity"`
	SoldQuantity int       `gorm:"type:int" json:"sold_quantity"`
	Stock        int       `gorm:"type:int" json:"stock"`
	IsStock      bool      `gorm:"type:boolean;default:true" json:"is_stock"`

	Variant *Variant `gorm:"foreignKey:VariantID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (m *Inventory) SetStock() {
	m.Stock = m.Quantity - m.SoldQuantity
	if m.Stock <= 5 {
		m.IsStock = false
	}
}
