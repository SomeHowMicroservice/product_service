package initialization

import (
	"fmt"

	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var allModels = []interface{}{
	&model.Category{},
	&model.Product{},
	&model.Color{},
	&model.Size{},
	&model.Variant{},
	&model.Inventory{},
	&model.Image{},
	&model.Tag{},
}

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s sslmode=%s channel_binding=%s",
		cfg.Database.DBHost,
		cfg.Database.DBName,
		cfg.Database.DBUser,
		cfg.Database.DBPassword,
		cfg.Database.DBSSLMode,
		cfg.Database.DBChannelBinding,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("kết nối PostgreSQL thất bại: %w", err)
	}

	if err := runAutoMigrations(db); err != nil {
		return nil, fmt.Errorf("chuyển dịch DB thất bại: %w", err)
	}

	return db, nil
}

func runAutoMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(allModels...); err != nil {
		return err
	}

	return nil
}
