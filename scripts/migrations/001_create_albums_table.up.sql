CREATE TABLE IF NOT EXISTS albums (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    year INTEGER NOT NULL CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE)),
    genre VARCHAR(100),
    condition VARCHAR(50) NOT NULL CHECK (condition IN ('mint', 'very good', 'good', 'fair', 'poor')),
    in_stock BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_albums_artist ON albums(artist);
CREATE INDEX IF NOT EXISTS idx_albums_in_stock ON albums(in_stock);
CREATE INDEX IF NOT EXISTS idx_albums_year ON albums(year);

INSERT INTO albums (id, title, artist, price, year, genre, condition, in_stock, created_at, updated_at) VALUES
('1', 'Blue Train', 'John Contrane', 56.99, 1957, 'Hard Bop', 'mint', true),
('2', 'Kind of Blue', 'Miles Davis', 65.99, 1959, 'Modal Jazz', 'very good', true),
('3', 'A Love Supreme', 'John Contrane', 45.50, 1965, 'Spiritual Jazz', 'good', true),
('4', 'Time Out', 'Dave Brubeck', 39.99, 1959, 'Cool Jazz', 'fair', false);


