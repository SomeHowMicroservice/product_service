package imagekit

import (
	"context"

	"github.com/SomeHowMicroservice/shm-be/product/common"
)

type ImageKitService interface {
	UploadFromBase64(ctx context.Context, req *common.Base64UploadRequest) (*common.UploadFileResponse, error)

	DeleteFile(ctx context.Context, fileID string) error
}