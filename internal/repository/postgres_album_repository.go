// Пакет для работы с хранилищами данных
package repository

import (
	"database/sql"
	"fmt"
	"go-music-shop/internal/domain/models"
	"log"
	"time"
)

// PostgresAlbumRepository - реализация репозитория для PostgreSQL
// Вместо хранения в памяти (как MemoryAlbumRepository) работает с реальной БД
type PostgresAlbumRepository struct {
	db *sql.DB // Подключение к базе данных PostgreSQL
}

// NewPostgresAlbumRepository - конструктор (фабричная функция)
// Принимает готовое подключение к БД и возвращает репозиторий
func NewPostgresAlbumRepository(db *sql.DB) *PostgresAlbumRepository {
	return &PostgresAlbumRepository{db: db}
}

// GetAll - получает ВСЕ альбомы из базы данных
func (r *PostgresAlbumRepository) GetAll() ([]domain.Album, error) {
	// SQL запрос для получения всех альбомов
	// $1, $2... - это placeholders для параметров (в этом запросе их нет)

	query := `SELECT id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at 
    		FROM albums ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}
	defer rows.Close() // Важно: закрываем соединение когда функция завершится

	var albums []domain.Album

	// rows.Next() последовательно перебирает все строки результата
	for rows.Next() {
		var album domain.Album

		// rows.Scan() заполняет поля структуры значениями из текущей строки
		// Важно: порядок и количество параметров должен совпадать с SELECT!
		err := rows.Scan(
			&album.ID,
			&album.Title,
			&album.Artist,
			&album.Price,
			&album.Year,
			&album.Genre,
			&album.Condition,
			&album.InStock,
			&album.CreatedAt,
			&album.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}

		albums = append(albums, album)
	}

	// Проверяем не было ли ошибок во время итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return albums, nil
}

// GetByID - находит ОДИН альбом по его ID
func (r *PostgresAlbumRepository) GetByID(id string) (*domain.Album, error) {
	query := `SELECT id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at 
    		FROM albums WHERE id = $1`

	var album domain.Album

	// QueryRow возвращает ТОЛЬКО ОДНУ строку (или ошибку)
	// .Scan сразу заполняет структуру из результата
	err := r.db.QueryRow(query, id).Scan( // Передаем id как параметр $1
		&album.ID,
		&album.Title,
		&album.Artist,
		&album.Price,
		&album.Year,
		&album.Genre,
		&album.Condition,
		&album.InStock,
		&album.CreatedAt,
		&album.UpdatedAt,
	)

	// Проверяем специальный тип ошибки "строка не найдена"
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("album not found")
	}
	// Проверяем другие ошибки
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	return &album, nil
}

// Create - создает НОВЫЙ альбом в базе данных
func (r *PostgresAlbumRepository) Create(album *domain.Album) error {
	query := `INSERT INTO albums (id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	// Заполняем технические поля которые не приходят от пользователя
	album.ID = generateID()
	album.CreatedAt = time.Now()
	album.UpdatedAt = time.Now()

	// db.Exec выполняет запрос НЕ возвращающий строки (INSERT, UPDATE, DELETE)
	// Передаем все 10 параметров в правильном порядке
	_, err := r.db.Exec(
		query,
		album.ID,
		album.Title,
		album.Artist,
		album.Price,
		album.Year,
		album.Genre,
		album.Condition,
		album.InStock,
		album.CreatedAt,
		album.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create album: %w", err)
	}

	log.Printf("Created album with ID: %s", album.ID)
	return nil
}

func (r *PostgresAlbumRepository) Update(album *domain.Album) error {
	query := `UPDATE albums SET title = $1, artist = $2, price = $3, year = $4, genre = $5, condition = $6, in_stock = $7, updated_at = $8
		WHERE id = $9`

	// Обновляем время последнего изменения
	album.UpdatedAt = time.Now()

	// db.Exec выполняет запрос НЕ возвращающий строки (INSERT, UPDATE, DELETE)
	// Передаем все параметры в правильном порядке
	result, err := r.db.Exec(
		query,
		album.Title,
		album.Artist,
		album.Price,
		album.Year,
		album.Genre,
		album.Condition,
		album.InStock,
		album.UpdatedAt,
		album.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}

	// Проверяем был ли вообще обновлен какой-либо альбом
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updating rows error: %w", err)
	}

	// Если ни одна строка не обновлена - значит альбом с таким ID не найден
	if rowsAffected == 0 {
		return fmt.Errorf("album with ID %s not found", album.ID)
	}

	log.Printf("Updated album with ID: %s", album.ID)
	return nil
}

func (r *PostgresAlbumRepository) Delete(id string) error {
	query := `DELETE FROM albums WHERE id = $1`

	// db.Exec выполняет запрос НЕ возвращающий строки (INSERT, UPDATE, DELETE)
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	// Проверяем сколько строк было удалено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleting rows error: %w", err)
	}

	// Если ни одна строка не удалена - значит альбом с таким ID не найден
	if rowsAffected == 0 {
		return fmt.Errorf("album with ID %s not found", id)
	}

	log.Printf("Deleted album with ID: %s", id)
	return nil
}

func (r *PostgresAlbumRepository) GetByArtist(artist string) ([]domain.Album, error) {
	query := `SELECT id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at 
    		FROM albums WHERE artist = $1
			ORDER BY year DESC`

	rows, err := r.db.Query(query, artist)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums by artist: %w", err)
	}
	defer rows.Close() // Важно: закрываем соединение когда функция завершится

	var albums []domain.Album

	// rows.Next() последовательно перебирает все строки результата
	for rows.Next() {
		var album domain.Album

		// rows.Scan() заполняет поля структуры значениями из текущей строки
		// Важно: порядок и количество параметров должен совпадать с SELECT!
		err := rows.Scan(
			&album.ID,
			&album.Title,
			&album.Artist,
			&album.Price,
			&album.Year,
			&album.Genre,
			&album.Condition,
			&album.InStock,
			&album.CreatedAt,
			&album.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}

		albums = append(albums, album)
	}

	// Проверяем не было ли ошибок во время итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return albums, nil
}

func (r *PostgresAlbumRepository) GetInStock() ([]domain.Album, error) {
	query := `SELECT id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at
	FROM albums WHERE in_stock = true
	ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}
	defer rows.Close()

	var albums []domain.Album

	for rows.Next() {
		var album domain.Album

		err := rows.Scan(
			&album.ID,
			&album.Title,
			&album.Artist,
			&album.Price,
			&album.Year,
			&album.Genre,
			&album.Condition,
			&album.InStock,
			&album.CreatedAt,
			&album.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan album: %w", err)
		}

		albums = append(albums, album)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return albums, nil
}
