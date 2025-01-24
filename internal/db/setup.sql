CREATE TABLE IF NOT EXISTS media (
    id SERIAL PRIMARY KEY,
    media_host_id VARCHAR(255) UNIQUE NOT NULL, -- ID assigned by media host
    title VARCHAR(255) NOT NULL,
    media_type VARCHAR(50) NOT NULL,
    media_image VARCHAR
);
CREATE INDEX idx_media_host_id ON media(media_host_id);

CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    time_of_media INT NOT NULL,
    media_id INT REFERENCES media(id) ON DELETE CASCADE,
    poster VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    created_at timestamp default (now() at time zone 'utc')
);

