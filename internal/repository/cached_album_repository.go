// Репозиторий с кэшированием (Decorator Pattern)
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"go-music-shop/internal/domain"
	"go-music-shop/pkg/redis"
	"log"
	"time"
)

// CachedAlbumRepository - декоратор, который добавляет кэширование к любому репозиторию
// Используем паттерн Decorator чтобы не изменять существующий код
type CachedAlbumRepository struct {
	repo    domain.AlbumRepository // Оригинальный репозиторий (PostgreSQL)
	redis   *redis.RedisClient     // Redis клиент для кэширования
	timeOut time.Duration          // Таймаут для операций с Redis
}

// NewCachedAlbumRepository - конструктор кэшированного репозитория
func NewCachedAlbumRepository(repo domain.AlbumRepository, redisClient *redis.RedisClient) *CachedAlbumRepository {
	return &CachedAlbumRepository{
		repo:    repo,
		redis:   redisClient,
		timeOut: 2 * time.Second, // 2 секунды таймаут для Redis операций
	}
}

// generateCacheKey - генерирует ключ для кэша на основе типа данных и ID
func (c *CachedAlbumRepository) generateCacheKey(dataType string, id string) string {
	return fmt.Sprintf("album:%s:%s", dataType, id)
}

// GetAll - получает все альбомы с кэшированием
func (c *CachedAlbumRepository) GetAll() ([]domain.Album, error) {
	cacheKey := c.generateCacheKey("all", "")

	// Создаем контекст с таймаутом для Redis
	ctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	// Пытаемся получить данные из кэша
	cachedData, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("reading from cache error: %v", err)
		// Продолжаем без кэша - получаем данные из базы
	}

	// Если данные есть в кэше - возвращаем их
	if cachedData != "" {
		var albums []domain.Album
		if err := json.Unmarshal([]byte(cachedData), &albums); err == nil {
			log.Println("data from cache has been delivered (all albums)")
			return albums, nil
		} else {
			log.Printf("parsing cached data error: %v", err)
		}
	}

	// Если данных нет в кэше - получаем из базы
	albums, err := c.repo.GetAll()
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш асинхронно (не блокируем ответ)
	go func() {
		ctx := context.Background()
		if data, err := json.Marshal(albums); err == nil {
			// Сохраняем на 1 минуту для списка всех альбомов
			if err := c.redis.Set(ctx, cacheKey, string(data), time.Minute); err != nil {
				log.Printf("saving in cache error: %v", err)
			} else {
				log.Println("data has been saved in cache (all albums)")
			}
		}
	}()

	return albums, nil
}

// GetByID - получает альбом по ID с кэшированием
func (c *CachedAlbumRepository) GetByID(id string) (*domain.Album, error) {
	cacheKey := c.generateCacheKey("id", id)

	// Создаем контекст с таймаутом для Redis
	ctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	// Пытаемся получить данные из кэша
	cachedData, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("reading from cache error: %v", err)
		// Продолжаем без кэша - получаем данные из базы
	}

	// Если данные есть в кэше - возвращаем их
	if cachedData != "" {
		var album domain.Album
		if err := json.Unmarshal([]byte(cachedData), &album); err == nil {
			log.Printf("data from cache has been delivered (album by id)")
			return &album, nil
		} else {
			log.Printf("parsing cache data error: %v", err)
		}
	}

	// Если данных нет в кэше - получаем из базы
	album, err := c.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш асинхронно (не блокируем ответ)
	go func() {
		ctx := context.Background()
		if data, err := json.Marshal(album); err == nil {
			// Сохраняем на 5 минут для отдельного альбома
			if err := c.redis.Set(ctx, cacheKey, string(data), 5*time.Minute); err != nil {
				log.Printf("saving in cache error: %v", err)
			} else {
				log.Println("data has been saved in cache (album by id)")
			}
		}
	}()

	return album, nil
}

// Create - создает альбом БЕЗ удаления кэша всех альбомов
func (c *CachedAlbumRepository) Create(album *domain.Album) error {
	// Просто создаем в базе
	err := c.repo.Create(album)
	if err != nil {
		return err
	}

	// Инвалидируем кэши, которые зависят от этого альбома
	go func() {
		c.invalidateCache("artist", album.Artist) // Кэш альбомов этого исполнителя
		c.invalidateCache("stock", "")            // Кэш альбомов в наличии
		c.cacheAlbum(album)                       // Кэшируем новый альбом
	}()

	return nil
}

// cacheAlbum - кэширует отдельный альбом
func (c *CachedAlbumRepository) cacheAlbum(album *domain.Album) {
	cacheKey := c.generateCacheKey("id", album.ID)
	ctx := context.Background()

	if data, err := json.Marshal(album); err == nil {
		if err := c.redis.Set(ctx, cacheKey, string(data), 5*time.Minute); err != nil {
			log.Printf("⚠️ Ошибка кэширования альбома: %v", err)
		} else {
			log.Printf("💾 Новый альбом %s закэширован", album.ID)
		}
	}
}

// Update - обновляет альбом и инвалидирует только его кэш
func (c *CachedAlbumRepository) Update(album *domain.Album) error {
	// Получаем старый альбом чтобы знать предыдущего исполнителя
	oldAlbum, _ := c.repo.GetByID(album.ID)

	err := c.repo.Update(album)
	if err != nil {
		return err
	}

	go func() {
		// Инвалидируем кэши связанные с альбомом
		c.invalidateCache("id", album.ID)

		if oldAlbum != nil {
			c.invalidateCache("artist", oldAlbum.Artist) // Старый исполнитель
		}

		c.invalidateCache("artist", album.Artist) // Новый исполнитель
		c.invalidateCache("stock", "")            // Кэш наличия

	}()

	return nil
}

// Delete - удаляет альбом и инвалидирует только его кэш
func (c *CachedAlbumRepository) Delete(id string) error {
	// Получаем альбом перед удалением чтобы знать исполнителя
	album, _ := c.repo.GetByID(id)

	err := c.repo.Delete(id)
	if err != nil {
		return err
	}

	go func() {
		c.invalidateCache("id", id)
		if album != nil {
			c.invalidateCache("artist", album.Artist) // Инвалидируем кэш исполнителя
		}
		c.invalidateCache("stock", "") // Инвалидируем кэш наличия
	}()

	return nil
}

// invalidateCache - удаляет данные из кэша
func (c *CachedAlbumRepository) invalidateCache(dataType string, id string) {
	cacheKey := c.generateCacheKey(dataType, id)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	if err := c.redis.Delete(ctx, cacheKey); err != nil {
		log.Printf("Ошибка инвалидации кэша %s: %v", cacheKey, err)
	}
}

func (c *CachedAlbumRepository) GetByArtist(artist string) ([]domain.Album, error) {
	cacheKey := c.generateCacheKey("artist", artist)

	// Создаем контекст с таймаутом для Redis
	ctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	// Пытаемся получить данные из кэша
	cachedData, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("reading from cache error: %v", err)
		// Продолжаем без кэша - получаем данные из базы
	}

	// Если данные есть в кэше - возвращаем их
	if cachedData != "" {
		var albums []domain.Album
		if err := json.Unmarshal([]byte(cachedData), &albums); err == nil {
			log.Printf("data from cache has been delivered (albums by artist %s)", artist)
			return albums, nil
		} else {
			log.Printf("parsing cache data error: %v", err)
		}
	}

	// Если данных нет в кэше - получаем из базы
	albums, err := c.repo.GetByArtist(artist)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш асинхронно (не блокируем ответ)
	go func() {
		ctx := context.Background()
		if data, err := json.Marshal(albums); err == nil {
			// Сохраняем на 2 минуты
			if err := c.redis.Set(ctx, cacheKey, string(data), 2*time.Minute); err != nil {
				log.Printf("saving in cache error: %v", err)
			} else {
				log.Printf("data has been saved in cache (albums by artist %s)", artist)
			}
		}
	}()

	return albums, nil
}

func (c *CachedAlbumRepository) GetInStock() ([]domain.Album, error) {
	cacheKey := c.generateCacheKey("stock", "")

	ctx, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	// Пытаемся получить из кеша
	cachedData, err := c.redis.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("reading from cache error: %v", err)
	}

	// Если данные есть в кэше - возвращаем их
	if cachedData != "" {
		var albums []domain.Album
		if err := json.Unmarshal([]byte(cachedData), &albums); err == nil {
			log.Printf("data from cache has been delivered (albums in stock)")
			return albums, nil
		} else {
			log.Printf("parsing from cache error: %v", err)
		}
	}

	// Если данных нет в кэше - загружаем из бд
	albums, err := c.repo.GetInStock()
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш асинхронно на 30 секунд (т.к часто меняются)
	go func() {
		ctx := context.Background()
		if data, err := json.Marshal(albums); err == nil {
			if err := c.redis.Set(ctx, cacheKey, string(data), 30*time.Second); err != nil {
				log.Printf("saving in cache error: %v", err)
			} else {
				log.Printf("data has been saved in cache (albums in stock)")
			}
		}
	}()

	return albums, nil
}
