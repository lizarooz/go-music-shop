package main

import (
	"go-music-shop/internal/catalog"
	"go-music-shop/internal/config"
	"go-music-shop/internal/repository"
	"go-music-shop/internal/service"
	"go-music-shop/pkg/database"
	"go-music-shop/pkg/redis"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// Импортируем сгенерированный код
	catalogpb "go-music-shop/pkg/gen/catalog"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к PostgreSQL
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("could not connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Подключаемся к Redis
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("could not connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Создаем репозитории
	postgresRepo := repository.NewPostgresAlbumRepository(db)
	cachedRepo := repository.NewCachedAlbumRepository(postgresRepo, redisClient)

	//Создаем СЕРВИСНЫЙ СЛОЙ (AlbumService)
	albumService := service.NewAlbumService(cachedRepo)

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()

	// Регистрируем наш сервис
	catalogService := catalog.NewCatalogService(albumService)
	catalogpb.RegisterCatalogServiceServer(grpcServer, catalogService)

	// Включаем reflection для тестирования (dev only)
	reflection.Register(grpcServer)

	// Запускаем gRPC сервер
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("starting gRPC server error: %v", err)
	}

	log.Printf("gRPC Catalog Service has been started on port :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}