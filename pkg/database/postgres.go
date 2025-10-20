// Пакет для работы с базой данных
package database

import (
	"database/sql"
	"fmt"
	"go-music-shop/internal/config"
	"time"
	_ "github.com/lib/pq"

)

// NewPostgresConnection - создает и настраивает подключение к PostgreSQL
// Принимает конфигурацию и возвращает готовое подключение или ошибку
func NewPostgresConnection(cfg *config.Config) (*sql.DB, error) {
	// Формируем строку подключения к PostgreSQL
	// Это специальный формат который понимает драйвер PostgreSQL
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DataBase.Host,
		cfg.DataBase.Port,
		cfg.DataBase.User,
		cfg.DataBase.Password,
		cfg.DataBase.Name,
		cfg.DataBase.SSLMode,
	)

	// sql.Open() создает объект подключения к БД
	// "postgres" - указываем какой драйвер использовать
	//  connStr - строка подключения с параметрами
	//  На этом этапе подключение еще не устанавливается!
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// НАСТРОЙКА ПУЛА ПОДКЛЮЧЕНИЙ - очень важная часть!

	// SetMaxOpenConns - максимальное количество ОДНОВРЕМЕННЫХ подключений к БД
    // Если все 25 подключений заняты - новые запросы будут ждать в очереди
	db.SetMaxOpenConns(25)

	// SetMaxIdleConns - количество подключений которые сохраняются "в запасе"
    // Эти подключения готовы к использованию без установки нового соединения
	db.SetMaxIdleConns(25)

	// SetConnMaxLifetime - максимальное время жизни подключения
    // Через 5 минут подключение закрывается и создается новое
    // Это помогает балансировать нагрузку и избегать "устаревших" подключений
	db.SetConnMaxLifetime(5 * time.Minute)

	// ТЕСТИРУЕМ ПОДКЛЮЧЕНИЕ
    // db.Ping() устанавливает реальное соединение с БД и проверяет что все работает
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
