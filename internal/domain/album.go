package domain

import "time"

// album represents data about a record album.
type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title" validate:"required"`
	Artist string  `json:"artist" validate:"required"`
	Price  float64 `json:"price" validate:"min=0"`
	Year int `json:"year"`
	Genre string `json:"genre"`
	Condition string `json:"condition"` // "mint", "very good", "good", "fair"
	InStock bool `json:"in_stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlbumRepository - интерфейс для работы с хранилищем альбомов.
// Это контракт, который должны реализовывать все репозитории
type AlbumRepository interface {
	GetAll() ([]Album, error)
	GetByID(id string) (*Album, error)
	Create(album *Album) error
	Update(album *Album) error
	Delete(id string) error
	GetByArtist(artist string) ([]Album, error)
	GetInStock()([]Album, error) // альбомы в наличии
}