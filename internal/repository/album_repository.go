// Реализация интерфейсов из domain слоя для работы с данными. Repository - работает с данными (CRUD операции)
package repository

import (
	"fmt"
	"go-music-shop/internal/domain"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

// MemoryAlbumRepository - in-memory реализация репозитория
type MemoryAlbumRepository struct {
	albums []domain.Album
	mu     sync.RWMutex
}

// NewMemoryAlbumRepository - конструктор репозитория
func NewMemoryAlbumRepository() *MemoryAlbumRepository {
	return &MemoryAlbumRepository{
		albums: []domain.Album{
			{
				ID:        "1",
				Title:     "Blue Train",
				Artist:    "John Coltrane",
				Price:     56.99,
				Year:      1957,
				Genre:     "Hard Bop",
				Condition: "mint",
				InStock:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			//TODO: ...
		}, // Явная инициализация mu: sync.RWMutex{} не требуется — она произойдет автоматически.
	}
}

// GetAll - возвращает все альбомы
func (r *MemoryAlbumRepository) GetAll() ([]domain.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.albums, nil
}

// GetByID - находит альбом по ID
func (r *MemoryAlbumRepository) GetByID(id string) (*domain.Album, error) {
	r.mu.RLock()         // Захватываем блокировку на чтение
	defer r.mu.RUnlock() // Гарантируем разблокировку при выходе из функции

	for _, album := range r.albums {
		if album.ID == id {
			return &album, nil
		}
	}

	return nil, fmt.Errorf("album not found")
}

// Create - добавляет новый альбом
func (r *MemoryAlbumRepository) Create(album *domain.Album) error {
	r.mu.Lock()         // Захватываем эксклюзивную блокировку на запись
	defer r.mu.Unlock() // Гарантируем разблокировку (Lock()/Unlock() вместо RLock()/RUnlock(), потому что мы изменяем данные (добавляем новый альбом в слайс).)

	album.ID = generateID()
	album.CreatedAt = time.Now()
	album.UpdatedAt = time.Now()

	r.albums = append(r.albums, *album)

	return nil
}

// Update - обновляет поля альбома
func (r *MemoryAlbumRepository) Update(album *domain.Album) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, a := range r.albums {
		if a.ID == album.ID {
			// Сохраняем CreatedAt из оригинала
			album.CreatedAt = a.CreatedAt
			album.UpdatedAt = time.Now()

			r.albums[i] = *album
			return nil
		}
	}

	return fmt.Errorf("album with ID %s not found", album.ID)
}

// Delete - удаляет альбом по ID
func (r *MemoryAlbumRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for idx, album := range r.albums {
		if album.ID == id {
			// Удаляем элемент из слайса
			r.albums = slices.Delete(r.albums, idx, idx+1)
			return nil
		}
	}
	return fmt.Errorf("album with ID %s not found", id)
}

// GetByArtist - находит альбом по автору
func (r *MemoryAlbumRepository) GetByArtist(artist string) ([]domain.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var albumsByArtist []domain.Album

	for _, album := range r.albums {
		if album.Artist == artist {
			albumsByArtist = append(albumsByArtist, album)
		}
	}

	if len(albumsByArtist) == 0 {
		return nil, fmt.Errorf("no albums found for artist %s", artist)
	}

	return albumsByArtist, nil
}

// GetInStock - проверяет в наличии ли альбом
func (r *MemoryAlbumRepository) GetInStock() ([]domain.Album, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var albumsInStock []domain.Album

	for _, album := range r.albums {
		if album.InStock {
			albumsInStock = append(albumsInStock, album)
		}
	}

	return albumsInStock, nil
}

// generateID - генерирует уникальный id
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
