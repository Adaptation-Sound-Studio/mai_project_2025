CREATE TABLE IF NOT EXISTS genres (
    genre_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(80) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS artists (
    artist_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS songs (
    song_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS fact_listens (
    fact_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    song_id BIGINT,
    artist_id BIGINT,
    genre_id BIGINT,
    listened_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_fact_listens_song FOREIGN KEY (song_id)
        REFERENCES songs (song_id) ON DELETE SET NULL,
    CONSTRAINT fk_fact_listens_artist FOREIGN KEY (artist_id)
        REFERENCES artists (artist_id) ON DELETE SET NULL,
    CONSTRAINT fk_fact_listens_genre FOREIGN KEY (genre_id)
        REFERENCES genres (genre_id) ON DELETE SET NULL
);

CREATE INDEX idx_fact_listens_user_listened_at
    ON fact_listens (user_id, listened_at);

-- Вставка жанров
INSERT INTO genres (name) VALUES
  ('Pop'),
  ('Rock'),
  ('Jazz');

-- Вставка артистов
INSERT INTO artists (name) VALUES
  ('Taylor Swift'),
  ('Queen'),
  ('Miles Davis');

-- Вставка песен
INSERT INTO songs (name) VALUES
  ('Love Story'),
  ('Bohemian Rhapsody'),
  ('So What');

-- Вставка фактов прослушиваний
INSERT INTO fact_listens (user_id, song_id, artist_id, genre_id, listened_at) VALUES
  (101, 1, 1, 1, '2025-06-01 10:00:00'),
  (102, 2, 2, 2, '2025-06-01 10:05:00'),
  (101, 3, 3, 3, '2025-06-01 10:10:00'),
  (103, 1, 1, 1, '2025-06-01 10:15:00');
