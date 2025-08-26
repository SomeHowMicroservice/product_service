package main

import (
	"fmt"
	"log"
	"net"

	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/consumers"
	"github.com/SomeHowMicroservice/shm-be/product/container"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	"github.com/SomeHowMicroservice/shm-be/product/initialization"
	productpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/product"
	"google.golang.org/grpc"
)

var (
	userAddr = "localhost:8082"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Tải cấu hình Product Service thất bại: %v", err)
	}

	db, err := initialization.InitDB(cfg)
	if err != nil {
		log.Fatalf("Lỗi kết nối DB ở Product Service: %v", err)
	}

	mqc, err := initialization.InitMessageQueue(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer mqc.Close()

	userAddr = cfg.App.ServerHost + fmt.Sprintf(":%d", cfg.Services.UserPort)
	clients := initialization.InitClients(userAddr)

	grpcServer := grpc.NewServer()
	productContainer := container.NewContainer(cfg, db, mqc.Chann, grpcServer, clients.UserClient)
	productpb.RegisterProductServiceServer(grpcServer, productContainer.GRPCHandler)

	imagekit := imagekit.NewImageKitService(cfg)
	go consumers.StartUploadImageConsumer(mqc, imagekit, productContainer.ImageRepo)
	go consumers.StartDeleteImageConsumer(mqc, imagekit)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe: %v", err)
	}
	defer lis.Close()

	log.Println("Khởi chạy service thành công")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Kết nối tới phục vụ thất bại: %v", err)
	}
}
