package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SomeHowMicroservice/shm-be/product/config"
	"github.com/SomeHowMicroservice/shm-be/product/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Tải cấu hình Product Service thất bại: %v", err)
	}

	server, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Khởi tạo service thất bại: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	errCh := make(chan error, 1)

	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	log.Println("Chạy service thành công")

	select {
		case <-stop:
			log.Println("Nhận được tín hiệu dừng server")
		case err := <- errCh:
			log.Printf("Chạy service thất bại: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
