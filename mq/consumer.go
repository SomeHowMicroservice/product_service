package mq

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	imageRepo "github.com/SomeHowMicroservice/shm-be/product/repository/image"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/bytedance/sonic"
)

var (
	uploadProgress = make(map[string]uint16)
	uploadTargets  = make(map[string]uint16)
	mu             sync.Mutex
)

func RegisterDeleteImageConsumer(router *message.Router, subscriber message.Subscriber, imagekit imagekit.ImageKitService) {
	router.AddConsumerHandler(
		"delete_image_handler",
		common.DeleteTopic,
		subscriber,
		message.NoPublishHandlerFunc(func(msg *message.Message) error {
			return handleDeleteImage(msg, imagekit)
		}),
	)
}

func RegisterUploadImageConsumer(router *message.Router, publisher message.Publisher, subscriber message.Subscriber, imagekit imagekit.ImageKitService, imageRepo imageRepo.ImageRepository) {
	router.AddHandler(
		"upload_image_handler",
		common.UploadTopic,
		subscriber,
		common.UploadedTopic,
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			var imageMsg *common.Base64UploadRequest
			if err := sonic.Unmarshal(msg.Payload, &imageMsg); err != nil {
				return nil, fmt.Errorf("unmarshal json thất bại: %w", err)
			}

			ctx := context.Background()
			res, err := imagekit.UploadFromBase64(ctx, imageMsg)
			if err != nil {
				return nil, fmt.Errorf("upload image thất bại: %w", err)
			}
			log.Printf("Tải lên hình ảnh thành công: %s", res.URL)

			fileID := res.FileID
			url := res.URL
			updateData := map[string]any{
				"file_id": fileID,
				"url":     url,
			}
			if err = imageRepo.Update(ctx, imageMsg.ImageID, updateData); err != nil {
				return nil, fmt.Errorf("cập nhật database thất bại: %w", err)
			}
			log.Printf("Cập nhật ảnh có FileID: %s và url: %s thành công", fileID, url)

			mu.Lock()
			uploadTargets[imageMsg.ProductID] = imageMsg.TotalImages
			uploadProgress[imageMsg.ProductID]++
			done := uploadProgress[imageMsg.ProductID] == uploadTargets[imageMsg.ProductID]
			mu.Unlock()

			log.Printf("Ảnh %s của product %s upload xong (%d/%d)", imageMsg.ImageID, imageMsg.ProductID, uploadProgress[imageMsg.ProductID], imageMsg.TotalImages)

			if done {
				event := common.ImageUploadedEvent{
					Service:   "product",
					UserID:    imageMsg.UserID,
					ProductID: imageMsg.ProductID,
				}
				body, _ := sonic.Marshal(event)
				out := message.NewMessage(watermill.NewUUID(), body)

				mu.Lock()
				delete(uploadProgress, imageMsg.ProductID)
				delete(uploadTargets, imageMsg.ProductID)
				mu.Unlock()

				return []*message.Message{out}, nil
			}

			return nil, nil
		},
	)
}

func handleDeleteImage(msg *message.Message, imagekit imagekit.ImageKitService) error {
	var imageMsg common.DeleteFileRequest
	if err := sonic.Unmarshal(msg.Payload, &imageMsg); err != nil {
		return fmt.Errorf("unmarshal json thất bại: %w", err)
	}

	ctx := context.Background()

	if err := imagekit.DeleteFile(ctx, imageMsg.FileID, imageMsg.FileUrl); err != nil {
		return fmt.Errorf("xóa file thất bại: %w", err)
	}

	log.Printf("Xóa hình ảnh có Url: %s thành công", imageMsg.FileUrl)
	return nil
}
