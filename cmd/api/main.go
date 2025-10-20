package main

import (
	"database/sql"
	"go-music-shop/internal/config"
	"go-music-shop/internal/handlers"
	"go-music-shop/internal/repository"
	"go-music-shop/internal/service"
	"go-music-shop/pkg/database"
	"go-music-shop/pkg/redis"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// config.Load() читает переменные окружения и возвращает структуру Config
	// Пример: ServerPort="8080", Database{Host:"localhost", Port:"5432", ...}
	cfg := config.Load()

	var db *sql.DB
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = database.NewPostgresConnection(cfg)
		if err == nil {
			log.Println("Successfully connected to PostgreSQL!")
			break
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
		return
	}
	defer db.Close()

	// Подключаемся к Redis
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Не удалось подлючиться к Redis: %v", err)
	}
	defer redisClient.Close()


	// Создаем цепочку зависимостей:

	// 1. Репозиторий - работает непосредственно с базой данных
	// Выполняет SQL запросы: SELECT, INSERT, UPDATE, DELETE
	postgresRepo := repository.NewPostgresAlbumRepository(db)

	cachedRepo := repository.NewCachedAlbumRepository(postgresRepo, redisClient)

	// 2. Сервис - содержит бизнес-логику приложения
	// Выполняет валидацию, проверки, бизнес-правила
	// Не знает о том, как хранятся данные (в памяти, в БД, в файле)
	albumService := service.NewAlbumService(cachedRepo)

	// 3. Обработчик - работает с HTTP запросами и ответами
	// Принимает JSON, возвращает JSON с правильными HTTP статусами
	albumHandler := handlers.NewAlbumHandler(albumService)

	router := gin.Default()

	// Регистрируем маршруты (URL пути) и связываем их с обработчиками
	router.GET("/albums", albumHandler.GetAlbums)
	router.GET("/albums/:id", albumHandler.GetAlbumByID)
	router.POST("/albums", albumHandler.CreateAlbum)
	router.PUT("/albums/:id", albumHandler.UpdateAlbum)
	router.DELETE("/albums/:id", albumHandler.DeleteAlbum)
	router.GET("/artists/:artist/albums", albumHandler.GetAlbumsByArtist)
	router.GET("/albums/stock", albumHandler.GetAlbumsInStock)

	// Маршрут для проверки здоровья приложения
	// Используется мониторингами чтобы проверить что приложение работает
	router.GET("/health", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "vintage-jazz-shop",
			"database": "connected",
			"redis": "connected",
		})
	})

	// Запускаем HTTP сервер на указанном порту
	log.Printf("Server starting on port %s", cfg.ServerPort)
	router.Run(":" + cfg.ServerPort)

}


