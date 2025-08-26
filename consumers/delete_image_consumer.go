package consumers

import (
	"context"
	"log"

	"github.com/SomeHowMicroservice/shm-be/product/common"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	"github.com/SomeHowMicroservice/shm-be/product/initialization"
	"github.com/SomeHowMicroservice/shm-be/product/mq"
)

func StartDeleteImageConsumer(mqc *initialization.MQConnection, imagekit imagekit.ImageKitService) {
	if err := mq.ConsumeMessage(mqc.Chann, common.DeleteQueueName, common.Exchange, common.DeleteRoutingKey, func(body []byte) error {
		fileID := string(body)
		ctx := context.Background()
		if err := imagekit.DeleteFile(ctx, fileID); err != nil {
			return err
		}

		log.Printf("Xóa hình ảnh có FileID: %s thành công", fileID)
		return nil
	}); err != nil {
		log.Printf("Lỗi khởi tạo delete image consumer: %v", err)
	}
}