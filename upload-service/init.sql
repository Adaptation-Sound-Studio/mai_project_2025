CREATE TABLE IF NOT EXISTS genres (
    genre_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(80) NOT NULL UNIQUE
);


CREATE TABLE IF NOT EXISTS artists (
  artist_id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  user_id BIGINT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS songs (
    song_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    auditions BIGINT NOT NULL DEFAULT 0,
    genre_id BIGINT NOT NULL,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    link VARCHAR(300) NOT NULL,
    CONSTRAINT fk_songs_genre FOREIGN KEY (genre_id) REFERENCES genres(genre_id)
);

CREATE TABLE albums (
    album_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    auditions BIGINT NOT NULL DEFAULT 0,
    artist_id BIGINT NOT NULL,
    genre_id BIGINT NOT NULL,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    CONSTRAINT fk_albums_artist FOREIGN KEY (artist_id) REFERENCES artists(artist_id),
    CONSTRAINT fk_albums_genre FOREIGN KEY (genre_id) REFERENCES genres(genre_id) 
);

CREATE TABLE song_album (
    sa_id BIGSERIAL PRIMARY KEY,
    song_id BIGINT NOT NULL,
    album_id BIGINT NOT NULL,
    CONSTRAINT fk_song_album_song FOREIGN KEY (song_id) REFERENCES songs(song_id),
    CONSTRAINT fk_song_album_album FOREIGN KEY (album_id) REFERENCES albums(album_id),
    CONSTRAINT uq_song_album UNIQUE (song_id, album_id)
);

CREATE TABLE song_artist (
    sa_id BIGSERIAL PRIMARY KEY,
    song_id BIGINT NOT NULL,
    artist_id BIGINT NOT NULL,
    CONSTRAINT fk_song_artist_song FOREIGN KEY (song_id) REFERENCES songs(song_id),
    CONSTRAINT fk_song_artist_artist FOREIGN KEY (artist_id) REFERENCES artists(artist_id),
    UNIQUE (song_id, artist_id)
);

INSERT INTO genres (name) VALUES 
  ('Pop'),
  ('Rock'),
  ('Jazz')
ON CONFLICT DO NOTHING;

INSERT INTO artists (name, user_id) VALUES 
  ('Artist One', 11),
  ('Artist Two', 22)
ON CONFLICT DO NOTHING;

INSERT INTO songs (name, auditions, genre_id, link) VALUES 
  ('Song A', 1000, 1, 'https://link-to-song-a.com'),
  ('Song B', 2000, 2, 'https://link-to-song-b.com')
ON CONFLICT DO NOTHING;

INSERT INTO albums (name, auditions, artist_id, genre_id) VALUES 
  ('Album X', 5000, 1, 1),
  ('Album Y', 3000, 2, 2)
ON CONFLICT DO NOTHING;

INSERT INTO song_album (song_id, album_id) VALUES 
  (1, 1),
  (2, 2)
ON CONFLICT DO NOTHING;

INSERT INTO song_artist (song_id, artist_id) VALUES 
  (1, 1),
  (2, 2)
ON CONFLICT DO NOTHING;