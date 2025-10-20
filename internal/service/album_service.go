// Содержит бизнес-логику приложения. Service - содержит бизнес-правила (валидация цены, наличия товара)
package service

import (
	"fmt"
	"go-music-shop/internal/domain"
)

// AlbumService - сервис для работы с альбомами
type AlbumService struct {
	repo domain.AlbumRepository
}

// NewAlbumService - конструктор сервиса
func NewAlbumService(repo domain.AlbumRepository) *AlbumService {
	return &AlbumService{repo: repo}
}

// GetAllAlbums - возвращает все альбомы
func (s *AlbumService) GetAllAlbums() ([]domain.Album, error) {
	return s.repo.GetAll()
}

// GetAlbumByID - возвращает альбом по ID
func (s *AlbumService) GetAlbumByID(id string) (*domain.Album, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	return s.repo.GetByID(id)
}

// CreateAlbum - создает новый альбом с валидацией
func (s *AlbumService) CreateAlbum(album *domain.Album) error {
	if album.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if album.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	return s.repo.Create(album)
}

// UpdateAlbum - обновляет поля альбома с валидацией
func (s *AlbumService) UpdateAlbum(album *domain.Album) error {
	if album.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if album.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if album.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	// Проверяем, существует ли альбом
	existingAlbum, err := s.repo.GetByID(album.ID)
	if err != nil {
		return fmt.Errorf("album not found %w", err)
	}
	
	// Сохраняем оригинальные поля, которые не должны меняться
	album.CreatedAt = existingAlbum.CreatedAt

	return s.repo.Update(album)
}	

// DeleteAlbum - удаляет альбом по ID
func (s *AlbumService) DeleteAlbum(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	return s.repo.Delete(id)
}

// GetAlbumsByArtist - возвращает альбомы по исполнителю
func (s *AlbumService) GetAlbumsByArtist(artist string) ([]domain.Album, error) {
	if artist == "" {
		return nil, fmt.Errorf("artist cannot be empty")
	}
	return s.repo.GetByArtist(artist) 
}

// GetAlbumsInStock - проверяет в наличии ли альбом
func (s *AlbumService) GetAlbumsInStock() ([]domain.Album, error) {
	return s.repo.GetInStock()
}