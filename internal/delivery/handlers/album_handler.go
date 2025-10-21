// Handler - работает с HTTP (форматы запросов, коды ответов)
package handlers

import (
	"go-music-shop/internal/domain/models"
	"go-music-shop/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlbumHandler struct {
	albumService *service.AlbumService
}

// NewAlbumHandler - конструктор обработчика
func NewAlbumHandler(albumService *service.AlbumService) *AlbumHandler {
	return &AlbumHandler{albumService: albumService}
}

// GetAlbums - обработчик для получения всех альбомов
func (h *AlbumHandler) GetAlbums(c *gin.Context) {
	albums, err := h.albumService.GetAllAlbums()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, albums)
}

// GetAlbumByID - обработчик для получения альбома по ID
func (h *AlbumHandler) GetAlbumByID(c *gin.Context) {
	id := c.Param("id")

	album, err := h.albumService.GetAlbumByID(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error":"album not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, album)
}

// CreateAlbum - обработчик для создания альбома
func (h *AlbumHandler) CreateAlbum(c *gin.Context) {
	var newAlbum domain.Album

	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error":"invalid input"})
		return
	}

	if err := h.albumService.CreateAlbum(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// UpdateAlbum - обработчик для обновления альбома
func (h *AlbumHandler) UpdateAlbum(c *gin.Context) {
	id := c.Param("id")
	
	var updatedAlbum domain.Album

	if err := c.BindJSON(&updatedAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Устанавливаем ID из URL параметра
	updatedAlbum.ID = id

	if err := h.albumService.UpdateAlbum(&updatedAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedAlbum)
}

// DeleteAlbum - обработчик для удаления альбома
func (h *AlbumHandler) DeleteAlbum(c *gin.Context) {
	id := c.Param("id")

	if err := h.albumService.DeleteAlbum(id); err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, nil) // 204 No Content для удаления
}

// GetAlbumsByArtist - обработчик для получения альбомов по автору
func (h *AlbumHandler) GetAlbumsByArtist(c *gin.Context) {
	artist := c.Param("artist")

	albums, err := h.albumService.GetAlbumsByArtist(artist)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, albums)
}

// GetAlbumsInStock - обработчик для получения альбомов по наличию
func (h *AlbumHandler) GetAlbumsInStock(c *gin.Context) {
	
	albums, err := h.albumService.GetAlbumsInStock()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "albums not found"})
		return
	}

	if len(albums) == 0 {
        c.JSON(http.StatusOK, []domain.Album{}) // Пустой массив вместо ошибки
        return
    }

	c.IndentedJSON(http.StatusOK, albums)
}
