package model

type Image struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	ProductID   string    `gorm:"type:char(36);not null" json:"-"`
	ColorID     string    `gorm:"type:char(36);not null" json:"-"`
	Url         string    `gorm:"type:varchar(255);not null" json:"url"`
	FileID      string    `gorm:"type:char(24)" json:"file_id"`
	SortOrder   int       `gorm:"type:int;not null" json:"sort_order"`
	IsThumbnail bool      `gorm:"type:boolean;not null" json:"is_thumbnail"`

	Product *Product `gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"product"`
	Color   *Color   `gorm:"foreignKey:ColorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"color"`
}
