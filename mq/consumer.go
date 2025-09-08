package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	imageRepo "github.com/SomeHowMicroservice/shm-be/product/repository/image"
	"github.com/ThreeDotsLabs/watermill/message"
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

func RegisterUploadImageConsumer(router *message.Router, subscriber message.Subscriber, imagekit imagekit.ImageKitService, imageRepo imageRepo.ImageRepository) {
	router.AddConsumerHandler(
		"upload_image_handler",
		common.UploadTopic,
		subscriber,
		message.NoPublishHandlerFunc(func(msg *message.Message) error {
			return handleUploadImage(msg, imagekit, imageRepo)
		}),
	)
}

func handleDeleteImage(msg *message.Message, imagekit imagekit.ImageKitService) error {
	fileID := string(msg.Payload)
	ctx := context.Background()

	if err := imagekit.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("xóa file thất bại: %w", err)
	}

	log.Printf("Xóa hình ảnh có FileID: %s thành công", fileID)
	return nil
}

func handleUploadImage(msg *message.Message, imagekit imagekit.ImageKitService, imageRepo imageRepo.ImageRepository) error {
	var imageMsg common.Base64UploadRequest
	if err := json.Unmarshal(msg.Payload, &imageMsg); err != nil {
		return fmt.Errorf("unmarshal json thất bại: %w", err)
	}

	ctx := context.Background()

	res, err := imagekit.UploadFromBase64(ctx, &imageMsg)
	if err != nil {
		return fmt.Errorf("upload image thất bại: %w", err)
	}
	log.Printf("Tải lên hình ảnh thành công: %s", res.URL)

	fileID := res.FileID
	url := res.URL
	updateData := map[string]interface{}{
		"file_id": fileID,
		"url":     url,
	}
	if err = imageRepo.Update(ctx, imageMsg.ImageID, updateData); err != nil {
		return fmt.Errorf("cập nhật database thất bại: %w", err)
	}
	log.Printf("Cập nhật ảnh có FileID: %s và url: %s thành công", fileID, url)

	return nil
}
