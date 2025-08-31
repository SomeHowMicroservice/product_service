package server

import (
	"time"

	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/container"
	"github.com/SomeHowMicroservice/shm-be/product/imagekit"
	productpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/product"
	userpb "github.com/SomeHowMicroservice/shm-be/product/protobuf/user"
	imageRepo "github.com/SomeHowMicroservice/shm-be/product/repository/image"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gorm.io/gorm"
)

type GRPCServer struct {
	Server    *grpc.Server
	ImageKit  imagekit.ImageKitService
	ImageRepo imageRepo.ImageRepository
}

func NewGRPCServer(cfg *config.Config, db *gorm.DB, mqChann *amqp091.Channel, userClient userpb.UserServiceClient) *GRPCServer {
	kaParams := keepalive.ServerParameters{
		Time:                  5 * time.Minute,
		Timeout:               20 * time.Second,
		MaxConnectionIdle:     0,
		MaxConnectionAge:      0,
		MaxConnectionAgeGrace: 0,
	}

	kaPolicy := keepalive.EnforcementPolicy{
		MinTime:             1 * time.Minute,
		PermitWithoutStream: true,
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(kaParams),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
	)

	productContainer := container.NewContainer(cfg, db, mqChann, grpcServer, userClient)

	productpb.RegisterProductServiceServer(grpcServer, productContainer.GRPCHandler)

	return &GRPCServer{
		grpcServer,
		productContainer.ImageKit,
		productContainer.ImageRepo,
	}
}
