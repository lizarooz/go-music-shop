package catalog

import (
	"context"
	"fmt"
	"go-music-shop/internal/domain/models"
	"go-music-shop/internal/service"
	"log"
	"time"

	// Импортируем сгенерированный protobuf код
	catalogpb "go-music-shop/pkg/gen/catalog"
)

// CatalogService реализует gRPC сервис для каталога
type CatalogService struct {
	catalogpb.UnimplementedCatalogServiceServer
	albumService *service.AlbumService
}

// NewCatalogService создает новый экземпляр CatalogService
func NewCatalogService(albumService *service.AlbumService) *CatalogService {
	return &CatalogService{
		albumService: albumService,
	}
}

// GetAlbums возвращает все альбомы (с пагинацией)
func (s *CatalogService) GetAlbums(ctx context.Context, req *catalogpb.GetAlbumsRequest) (*catalogpb.GetAlbumsResponse, error) {
	log.Printf("gRPC GetAlbums has been called: limit=%d, offset=%d", req.GetLimit(), req.GetOffset())

	// Получаем все альбомы из репозитория
	albums, err := s.albumService.GetAllAlbums()
	if err != nil {
		return nil, fmt.Errorf("could not get albums %v", err)
	}

	// Применяем пагинацию
	start := int(req.GetOffset())
	if start < 0 {
		start = 0
	}
	if start > len(albums) {
		start = len(albums)
	}

	end := start + int(req.GetLimit())
	if end > len(albums) {
		end = len(albums)
	}
	if req.GetLimit() == 0 {
		end = len(albums) // Если limit не указан, возвращаем все
	}

	paginatedAlbums := albums[start:end]

	// Конвертируем domain альбомы в protobuf альбомы
	pbAlbums := make([]*catalogpb.Album, len(paginatedAlbums))
	for i := range paginatedAlbums {
		pbAlbums[i] = s.domainToProtoAlbum(&paginatedAlbums[i])
	}

	log.Printf("%d albums had been returned (all: %d)", len(pbAlbums), len(albums))

	return &catalogpb.GetAlbumsResponse{
		Albums:     pbAlbums,
		TotalCount: int32(len(albums)),
	}, nil
}

// GetAlbumByID возвращает альбом по ID
func (s *CatalogService) GetAlbumByID(ctx context.Context, req *catalogpb.GetAlbumByIDRequest) (*catalogpb.GetAlbumByIDResponse, error) {
	id := req.GetId()
	log.Printf("gRPC GetAlbumByID has been called: id=%s", id)

	album, err := s.albumService.GetAlbumByID(id)
	if err != nil {
		return nil, fmt.Errorf("album not found: %w", err)
	}

	log.Printf("album was found: %s - %s", album.Artist, album.Title)

	return &catalogpb.GetAlbumByIDResponse{
		Album: s.domainToProtoAlbum(album),
	}, nil

}

// CreateAlbum создает новый альбом
func (s *CatalogService) CreateAlbum(ctx context.Context, req *catalogpb.CreateAlbumRequest) (*catalogpb.CreateAlbumResponse, error) {
	log.Printf("gRPC CreateAlbum has been called: %s - %s", req.GetArtist(), req.GetTitle())

	// Создаем domain альбом из запроса
	album := &domain.Album{
		Title:     req.GetTitle(),
		Artist:    req.GetArtist(),
		Price:     req.GetPrice(),
		Year:      int(req.GetYear()),
		Genre:     req.GetGenre(),
		Condition: req.GetCondition(),
		InStock:   req.GetInStock(),
	}

	if err := s.albumService.CreateAlbum(album); err != nil {
		return nil, fmt.Errorf("could not create album: %w", err)
	}

	log.Printf("album has been created: ID=%s", album.ID)

	return &catalogpb.CreateAlbumResponse{
		Album: s.domainToProtoAlbum(album),
	}, nil
}

// UpdateAlbum обновляет альбом
func (s *CatalogService) UpdateAlbum(ctx context.Context, req *catalogpb.UpdateAlbumRequest) (*catalogpb.UpdateAlbumResponse, error) {
	log.Printf("gRPC UpdateAlbum has been called: id=%s", req.GetId())

	// Создаем domain альбом из запроса
	album := &domain.Album{
		ID:        req.GetId(),
		Title:     req.GetTitle(),
		Artist:    req.GetArtist(),
		Price:     req.GetPrice(),
		Year:      int(req.GetYear()),
		Genre:     req.GetGenre(),
		Condition: req.GetCondition(),
		InStock:   req.GetInStock(),
	}

	if err := s.albumService.UpdateAlbum(album); err != nil {
		return nil, fmt.Errorf("could not update album: %w", err)
	}

	log.Printf("album has been updated: ID=%s", album.ID)

	return &catalogpb.UpdateAlbumResponse{
		Album: s.domainToProtoAlbum(album),
	}, nil
}

// DeleteAlbum удаляет альбом
func (s *CatalogService) DeleteAlbum(ctx context.Context, req *catalogpb.DeleteAlbumRequest) (*catalogpb.DeleteAlbumResponse, error) {
	id := req.GetId()
	log.Printf("gRPC DeleteAlbum has been called: id=%s", id)

	if err := s.albumService.DeleteAlbum(id); err != nil {
		return nil, fmt.Errorf("could not delete album: %w", err)
	}

	log.Printf("album has been deleted: ID=%s", id)

	return &catalogpb.DeleteAlbumResponse{
		Success: true,
		Message: "album has been deleted successfully",
	}, nil
}

// SearchAlbumsByArtist ищет альбомы по исполнителю
func (s *CatalogService) SearchAlbumsByArtist(ctx context.Context, req *catalogpb.SearchAlbumsByArtistRequest) (*catalogpb.SearchAlbumsByArtistResponse, error) {
	artist := req.GetArtist()
	log.Printf("gRPC SearchAlbumsByArtist has been called: artist=%s", artist)

	albums, err := s.albumService.GetAlbumsByArtist(artist) 
	if err != nil {
		return nil, fmt.Errorf("could not search albums: %w", err)
	}

	log.Printf("albums has been searched: artist=%s", artist)

	// Конвертируем domain альбомы в protobuf альбомы
	pbAlbums := make([]*catalogpb.Album, len(albums))
	for i, album := range albums {
		pbAlbums[i] = s.domainToProtoAlbum(&album)
	}

	return &catalogpb.SearchAlbumsByArtistResponse{
		Albums: pbAlbums,
	}, nil

}

// GetAlbumsInStock возвращает альбомы в наличии
func (s *CatalogService) GetAlbumsInStock(ctx context.Context, req *catalogpb.GetAlbumsInStockRequest) (*catalogpb.GetAlbumsInStockResponse, error) {
	log.Printf("gRPC GetAlbumsInStock has been called")

	albums, err := s.albumService.GetAlbumsInStock() 
	if err != nil {
		return nil, fmt.Errorf("could not search albums in stock: %w", err)
	}

	log.Printf("albums in stock has been searched")

	// Конвертируем domain альбомы в protobuf альбомы
	pbAlbums := make([]*catalogpb.Album, len(albums))
	for i, album := range albums {
		pbAlbums[i] = s.domainToProtoAlbum(&album)
	}

	return &catalogpb.GetAlbumsInStockResponse{
		Albums: pbAlbums,
	}, nil
}

// domainToProtoAlbum конвертирует domain.Album в catalogpb.Album
func (s *CatalogService) domainToProtoAlbum(album *domain.Album) *catalogpb.Album {
	return &catalogpb.Album{
		Id:        album.ID,
		Title:     album.Title,
		Artist:    album.Artist,
		Price:     album.Price,
		Year:      int32(album.Year),
		Genre:     album.Genre,
		Condition: album.Condition,
		InStock:   album.InStock,
		CreatedAt: album.CreatedAt.Format(time.RFC3339),
		UpdatedAt: album.UpdatedAt.Format(time.RFC3339),
	}
}
