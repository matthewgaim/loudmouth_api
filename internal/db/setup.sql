CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    profile_picture VARCHAR,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'utc')
);

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
    poster INT REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at timestamp default (now() at time zone 'utc')
);

