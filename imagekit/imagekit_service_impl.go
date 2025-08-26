package imagekit

import (
	"context"
	"fmt"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/uploader"
)

type imageKitServiceImpl struct {
	client *imagekit.ImageKit
}

func NewImageKitService(cfg *config.Config) ImageKitService {
	client := imagekit.NewFromParams(imagekit.NewParams{
		PrivateKey:  cfg.ImageKit.PrivateKey,
		PublicKey:   cfg.ImageKit.PublicKey,
		UrlEndpoint: cfg.ImageKit.URLEndpoint,
	})

	return &imageKitServiceImpl{client}
}

func (s *imageKitServiceImpl) UploadFromBase64(ctx context.Context, req *common.Base64UploadRequest) (*common.UploadFileResponse, error) {
	params := uploader.UploadParam{
		FileName:          req.FileName,
		UseUniqueFileName: boolPtr(false),
	}

	if req.Folder != "" {
		params.Folder = req.Folder
	}

	result, err := s.client.Uploader.Upload(ctx, req.Base64Data, params)
	if err != nil {
		return nil, fmt.Errorf("upload file thất bại: %w", err)
	}

	return &common.UploadFileResponse{
		FileID:       result.Data.FileId,
		Name:         result.Data.Name,
		URL:          result.Data.Url,
		ThumbnailURL: result.Data.ThumbnailUrl,
		Size:         int64(result.Data.Size),
		Width:        result.Data.Width,
		Height:       result.Data.Height,
	}, nil
}

func (s *imageKitServiceImpl) DeleteFile(ctx context.Context, fileID string) error {
	if _, err := s.client.Media.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("xóa file thất bại: %w", err)
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
