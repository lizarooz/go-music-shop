package redis

import (
	"context"
	"fmt"
	"go-music-shop/internal/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9" // Импортируем Redis клиент
)

// RedisClient - обертка вокруг Redis клиента с дополнительными методами
type RedisClient struct {
	client *redis.Client
	ttl    time.Duration // Время жизни кэша по умолчанию
}

// NewRedisClient - создает новое подключение к Redis
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	// Создаем опции для подключения к Redis
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Создаем клиент Redis
	client := redis.NewClient(options)

	// Проверяем подключение с помощью Ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis!")

	return &RedisClient{
		client: client,
		ttl:    time.Duration(cfg.Redis.DefaultTTL) * time.Second,
	}, nil
}

// Set - сохранение в кэш
func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// Если TTL не указан - используем значение по умолчанию
	if ttl == 0 {
		ttl = r.ttl
	}

	// Сохраняем в Redis: SET key value EX seconds
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("saving in Redis error: %w", err)
	}
	return nil
}

// Get - чтение из кэша
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()

	// Особенность: если ключ не найден - это НЕ ошибка для кэша
	if err == redis.Nil {
		return "", nil // Ключ не найден - нормально для кэша
	} else if err != nil {
		return "", fmt.Errorf("getting from Redis error: %w", err)
	}
	return value, nil
}

// Delete - удаление из кэша
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("deleting from Redis error: %w", err)
	}
	return nil
}

// Close - закрытие подключения
func (r *RedisClient) Close() error {
	// Закрываем подключение к Redis
	return r.client.Close()
}