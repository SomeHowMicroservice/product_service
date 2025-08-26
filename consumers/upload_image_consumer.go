package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	"github.com/SomeHowMicroservice/shm-be/product/initialization"
	"github.com/SomeHowMicroservice/shm-be/product/mq"
	imageRepo "github.com/SomeHowMicroservice/shm-be/product/repository/image"
)

func StartUploadImageConsumer(mqc *initialization.MQConnection, imagekit imagekit.ImageKitService, imageRepo imageRepo.ImageRepository) {
	if err := mq.ConsumeMessage(mqc.Chann, common.UploadQueueName, common.Exchange, common.UploadRoutingKey, func(body []byte) error {
		var imageMsg common.Base64UploadRequest
		if err := json.Unmarshal(body, &imageMsg); err != nil {
			return fmt.Errorf("unmarshal json thất bại: %w", err)
		}
		
		ctx := context.Background()
		res, err := imagekit.UploadFromBase64(ctx, &imageMsg)
		if err != nil {
			return err
		}
		log.Printf("Tải lên hình ảnh thành công: %s", res.URL)

		fileID := res.FileID
		if err = imageRepo.UpdateFileID(ctx, imageMsg.ImageID, fileID); err != nil {
			return err
		}
		log.Printf("Cập nhật ảnh có FileID: %s thành công", fileID)
		
		return nil
	}); err != nil {
		log.Printf("Lỗi khởi tạo upload image consumer: %v", err)
	}
}