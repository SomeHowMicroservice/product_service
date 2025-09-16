package container

import (
	"github.com/SomeHowMicroservice/product/config"
	"github.com/SomeHowMicroservice/product/handler"
	"github.com/SomeHowMicroservice/product/imagekit"
	userpb "github.com/SomeHowMicroservice/product/protobuf/user"
	categoryRepo "github.com/SomeHowMicroservice/product/repository/category"
	colorRepo "github.com/SomeHowMicroservice/product/repository/color"
	imageRepo "github.com/SomeHowMicroservice/product/repository/image"
	inventoryRepo "github.com/SomeHowMicroservice/product/repository/inventory"
	productRepo "github.com/SomeHowMicroservice/product/repository/product"
	sizeRepo "github.com/SomeHowMicroservice/product/repository/size"
	tagRepo "github.com/SomeHowMicroservice/product/repository/tag"
	variantRepo "github.com/SomeHowMicroservice/product/repository/variant"
	"github.com/SomeHowMicroservice/product/service"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	GRPCHandler *handler.GRPCHandler
	ImageRepo   imageRepo.ImageRepository
	ImageKit    imagekit.ImageKitService
}

func NewContainer(cfg *config.Config, db *gorm.DB, publisher message.Publisher, grpcServer *grpc.Server, userClient userpb.UserServiceClient) *Container {
	imageKit := imagekit.NewImageKitService(cfg)
	categoryRepo := categoryRepo.NewCategoryRepository(db)
	productRepo := productRepo.NewProductRepository(db)
	tagRepo := tagRepo.NewTagRepository(db)
	colorRepo := colorRepo.NewColorRepository(db)
	sizeRepo := sizeRepo.NewSizeRepository(db)
	variantRepo := variantRepo.NewVariantRepository(db)
	inventoryRepo := inventoryRepo.NewInventoryRepository(db)
	imageRepo := imageRepo.NewImageRepository(db)
	svc := service.NewProductService(cfg, db, userClient, publisher, categoryRepo, productRepo, tagRepo, colorRepo, sizeRepo, variantRepo, inventoryRepo, imageRepo)
	hdl := handler.NewGRPCHandler(grpcServer, svc)
	return &Container{
		hdl,
		imageRepo,
		imageKit,
	}
}
