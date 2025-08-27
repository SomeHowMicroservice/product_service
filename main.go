package main

import (
	"fmt"
	"log"
	"net"

	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/consumers"
	"github.com/SomeHowMicroservice/shm-be/product/container"
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
	defer db.Close()

	mqc, err := initialization.InitMessageQueue(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer mqc.Close()

	userAddr = fmt.Sprintf("%s:%d", cfg.App.ServerHost, cfg.Services.UserPort)
	clients, err := initialization.InitClients(userAddr)
	if err != nil {
		log.Fatalf("Kết nối tới các dịch vụ khác thất bại: %v", err)
	}
	defer clients.Close()

	grpcServer := grpc.NewServer()
	productContainer := container.NewContainer(cfg, db.Gorm, mqc.Chann, grpcServer, clients.UserClient)
	productpb.RegisterProductServiceServer(grpcServer, productContainer.GRPCHandler)

	go consumers.StartUploadImageConsumer(mqc, productContainer.ImageKit, productContainer.ImageRepo)
	go consumers.StartDeleteImageConsumer(mqc, productContainer.ImageKit)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe: %v", err)
	}
	defer lis.Close()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Serve gRPC thất bại: %v", err)
	}

	log.Println("Khởi chạy service thành công")
}
