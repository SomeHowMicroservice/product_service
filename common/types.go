package common

import "gorm.io/gorm"

type Base64UploadRequest struct {
	ImageID    string `json:"image_id"`
	Base64Data string `json:"base64_data"`
	FileName   string `json:"file_name"`
	Folder     string `json:"folder"`
}

type UploadFileResponse struct {
	FileID       string `json:"file_id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Size         int64  `json:"size"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type Preload struct {
	Relation string
	Scope    func(*gorm.DB) *gorm.DB
}

type Locking struct {
	Strength string
	Options  string
}

type PaginationQuery struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Sort       string `json:"sort"`
	Order      string `json:"order"`
	IsActive   *bool  `json:"is_active"`
	Search     string `json:"search"`
	CategoryID string `json:"category_id"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}
