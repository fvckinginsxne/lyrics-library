CREATE TABLE IF NOT EXISTS songs
(
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    artist VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    lyrics TEXT[] NOT NULL,
    translation TEXT[] NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_songs_artist_lower ON songs (artist);
CREATE INDEX IF NOT EXISTS idx_songs_title_lower ON songs (title);
CREATE INDEX IF NOT EXISTS idx_songs_uuid ON songs (uuid);