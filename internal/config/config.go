// Этот пакет отвечает за конфигурацию приложения
package config

import (
	"os"
	"strconv"
)

// Config - главная структура конфигурации всего приложения
// Хранит ВСЕ настройки приложения в одном месте
type Config struct {
	ServerPort string
	DataBase DataBaseConfig
	Redis RedisConfig
}

// DatabaseConfig - структура для настроек конкретно базы данных
type DataBaseConfig struct {
	Host string
	Port string
	User string
	Password string
	Name string
	SSLMode string
}

// RedisConfig - структура для настроек Redis
type RedisConfig struct {
	Host string
	Port string
	Password string
	DB int // Номер базы данных Redis (0-15)
	// TTL - Time To Live (время жизни кэша в секундах)
	DefaultTTL int // Стандартное время жизни кэшированных данных
}

// Load - главная функция которая загружает всю конфигурацию
// Возвращает готовый объект Config со всеми настройками
func Load() *Config {
	return &Config{
		// Загружаем порт сервера из переменных окружения
        // Если переменной нет - используем "8080" по умолчанию
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Инициализируем настройки базы данных
		DataBase: DataBaseConfig{
			Host: getEnv("DB_HOST", "localhost"),
			Port: getEnv("DB_PORT", "5432"),
			User: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name: getEnv("DB_NAME", "jazz_shop"),
			SSLMode: getEnv("DB_SSL_MODE", "disable"),
		},

		// ДОБАВЛЯЕМ настройки Redis
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB: getEnvAsInt("REDIS_DB", 0),
			DefaultTTL: getEnvAsInt("REDIS_DEFAULT_TTL", 300), // 5 минут по умолчанию
		},
	}
}

// getEnv - вспомогательная функция для получения переменных окружения
func getEnv(key, defaultValue string) string {
	// os.Getenv(key) пытается получить значение переменной окружения
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// getEnvAsInt - аналогично getEnv, но преобразует значение в число
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != ""	{
		if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
}
	return defaultValue
}